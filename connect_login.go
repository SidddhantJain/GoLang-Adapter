func (c *LocalConnect) Login(
	apiToken string,
	apiSecret string,
	totp *string,
) error {
	if apiToken == "" || apiSecret == "" {
		return errors.New("invalid api_token or api_secret")
	}
	extraHeaders := map[string]string{
		"api_secret": apiSecret,
	}
	route := fmt.Sprintf("auth/realms/debroking/dsbpkc/login/%s", apiToken)
	r, err := c.sendRequest(
		"https://signin.definedgebroking.com",
		route,
		"GET",
		nil,
		nil,
		nil,
		nil,
		extraHeaders,
	)
	if err != nil {
		return err
	}

	otpToken, ok := r["otp_token"].(string)
	if !ok {
		return errors.New("failed to obtain otp_token")
	}

	// Get OTP/TOTP for 2FA
	var otp string
	if totp == nil {
		fmt.Print("Enter OTP/External TOTP: ")
		_, err := fmt.Scan(&otp)
		if err != nil {
			return errors.New("no OTP/TOTP provided")
		}
	} else {
		otp = *totp
	}

	// Compute the session key
	ac := sha256.New()
	ac.Write([]byte(otpToken + otp + apiSecret))
	acHex := hex.EncodeToString(ac.Sum(nil))

	// Get session keys
	r, err = c.sendRequest(
		"https://signin.definedgebroking.com",
		"auth/realms/debroking/dsbpkc/token",
		"POST",
		nil,
		map[string]interface{}{
			"otp_token": otpToken,
			"otp":       otp,
			"ac":        acHex,
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return err
	}

	// Set session keys
	c.setSessionKeys(r["uid"].(string), r["actid"].(string), r["api_session_key"].(string), r["susertoken"].(string))

	// Attempt to remove symbols file
	symbolsFilename := filepath.Join(filepath.Dir(os.Args[0]), "allmaster.csv")
	if err := os.Remove(symbolsFilename); err != nil && !os.IsNotExist(err) {
		return err
	}
	c.Symbols = make(chan map[string]interface{}, 1)

	select {
	case c.Symbols <- map[string]interface{}{}:
	default:
	}

	return nil
}
