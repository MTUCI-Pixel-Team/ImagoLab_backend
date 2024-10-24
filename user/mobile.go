package user

import (
	"RestAPI/core"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	recaptcha "cloud.google.com/go/recaptchaenterprise/v2/apiv1"
	recaptchapb "cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
)

const FIREBASE_URL string = "https://identitytoolkit.googleapis.com/v1/accounts:sendVerificationCode?key="

type VerificationCodeRequest struct {
	PhoneNum       string `json:"phoneNumber"`
	RecaptchaToken string `json:"recaptchaToken"`
	Message        string `json:"message"`
}

func SendSMS(phone, recaptchaToken string, code int) error {
	if code < 100000 || code > 999999 {
		return fmt.Errorf("Invalid code: %d", code)
	}
	if phone == "" {
		return fmt.Errorf("Mobile number is empty")
	}
	if recaptchaToken == "" {
		return fmt.Errorf("Recaptcha token is empty")
	}

	msg := fmt.Sprintf("Your verification code: %d", code)
	url := FIREBASE_URL + core.FIREBASE_API_KEY
	req := VerificationCodeRequest{
		PhoneNum:       phone,
		RecaptchaToken: recaptchaToken,
		Message:        msg,
	}

	requestData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestData))
	if err != nil {
		return fmt.Errorf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to send code: %v", err)
	}

	return nil
}

func getRecaptchaToken() string {
	token := "action-token"
	recaptchaAction := "action-name"

	ctx := context.Background()
	client, err := recaptcha.NewClient(ctx)
	if err != nil {
		fmt.Printf("Error creating reCAPTCHA client\n")
	}
	defer client.Close()

	event := &recaptchapb.Event{
		Token:   token,
		SiteKey: core.RECAPTCHA_KEY,
	}

	assessment := &recaptchapb.Assessment{
		Event: event,
	}

	request := &recaptchapb.CreateAssessmentRequest{
		Assessment: assessment,
		Parent:     fmt.Sprintf("projects/%s", core.PROJECT_ID),
	}

	response, err := client.CreateAssessment(
		ctx,
		request)

	if err != nil {
		fmt.Printf("Error calling CreateAssessment: %v", err.Error())
	}

	// Check if the token is valid.
	if !response.TokenProperties.Valid {
		fmt.Printf("The CreateAssessment() call failed because the token was invalid for the following reasons: %v",
			response.TokenProperties.InvalidReason)
		return ""
	}

	// Check if the expected action was executed.
	if response.TokenProperties.Action != recaptchaAction {
		fmt.Printf("The action attribute in your reCAPTCHA tag does not match the action you are expecting to score")
		return ""
	}

	// Get the risk score and the reason(s).
	// For more information on interpreting the assessment, see:
	// https://cloud.google.com/recaptcha-enterprise/docs/interpret-assessment
	fmt.Printf("The reCAPTCHA score for this token is:  %v", response.RiskAnalysis.Score)
	return token
}
