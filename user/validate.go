package user

import (
	"RestAPI/db"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

type ValidationRules struct {
	EmailRegex    *regexp.Regexp
	UsernameRegex *regexp.Regexp
	PasswordRules PasswordRules
}

type PasswordRules struct {
	MinLength         int
	RequireUppercase  bool
	RequireLowercase  bool
	RequireDigits     bool
	RequireSpecial    bool
	DisallowedStrings []string
}

func DefaultValidationRules() *ValidationRules {
	return &ValidationRules{
		EmailRegex:    regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`),
		UsernameRegex: regexp.MustCompile(`^[a-zA-Z0-9_\-.]{4,64}$`),
		PasswordRules: PasswordRules{
			MinLength:         8,
			RequireUppercase:  true,
			RequireLowercase:  true,
			RequireDigits:     true,
			RequireSpecial:    true,
			DisallowedStrings: []string{"password", "12345678", "qwerty"},
		},
	}
}

func ValidateUser(user *db.User, keepFields ...[]string) (*db.User, error) {
	rules := DefaultValidationRules()
	var errors ValidationErrors

	if err := ValidateEmail(user.Email, rules); err != nil {
		errors = append(errors, *err)
	}

	if err := ValidateUsername(user.Username, rules); err != nil {
		errors = append(errors, *err)
	}

	if err := ValidatePassword(user.Password, rules); err != nil {
		errors = append(errors, *err)
	}

	if len(errors) > 0 {
		return user, errors
	}

	if len(keepFields) > 0 {
		validUser := FilterFields(user, keepFields[0])
		return validUser, nil
	}

	return user, nil
}

func ValidateEmail(email string, rules *ValidationRules) *ValidationError {
	if email == "" {
		return &ValidationError{
			Field:   "email",
			Message: "email is required",
		}
	}

	if !rules.EmailRegex.MatchString(email) {
		return &ValidationError{
			Field:   "email",
			Message: "invalid email format",
		}
	}

	if len(email) > 254 {
		return &ValidationError{
			Field:   "email",
			Message: "email is too long",
		}
	}

	if !CheckDomain(email) {
		return &ValidationError{
			Field:   "email",
			Message: "email domain is not allowed",
		}
	}

	return nil
}

func ValidateUsername(username string, rules *ValidationRules) *ValidationError {
	if username == "" {
		return &ValidationError{
			Field:   "username",
			Message: "username is required",
		}
	}

	if !rules.UsernameRegex.MatchString(username) {
		return &ValidationError{
			Field:   "username",
			Message: "username must be 4-64 characters long and contain only letters, numbers and underscores",
		}
	}

	return nil
}

func ValidatePassword(password string, rules *ValidationRules) *ValidationError {
	if password == "" {
		return &ValidationError{
			Field:   "password",
			Message: "password is required",
		}
	}

	if len(password) < rules.PasswordRules.MinLength {
		return &ValidationError{
			Field:   "password",
			Message: fmt.Sprintf("password must be at least %d characters long", rules.PasswordRules.MinLength),
		}
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if rules.PasswordRules.RequireUppercase && !hasUpper {
		return &ValidationError{
			Field:   "password",
			Message: "password must contain at least one uppercase letter",
		}
	}

	if rules.PasswordRules.RequireLowercase && !hasLower {
		return &ValidationError{
			Field:   "password",
			Message: "password must contain at least one lowercase letter",
		}
	}

	if rules.PasswordRules.RequireDigits && !hasDigit {
		return &ValidationError{
			Field:   "password",
			Message: "password must contain at least one digit",
		}
	}

	if rules.PasswordRules.RequireSpecial && !hasSpecial {
		return &ValidationError{
			Field:   "password",
			Message: "password must contain at least one special character",
		}
	}

	passwordLower := strings.ToLower(password)
	for _, disallowed := range rules.PasswordRules.DisallowedStrings {
		if strings.Contains(passwordLower, strings.ToLower(disallowed)) {
			return &ValidationError{
				Field:   "password",
				Message: "password is too common or contains forbidden patterns",
			}
		}
	}

	return nil
}

func FilterFields(user *db.User, fields []string) *db.User {
	if user == nil {
		return nil
	}

	if len(fields) == 0 {
		return user
	}

	v := reflect.ValueOf(user).Elem()
	t := v.Type()

	keep := make(map[string]bool)
	for _, field := range fields {
		for i := 0; i < v.NumField(); i++ {
			if t.Field(i).Name == field {
				keep[field] = true
				break
			}
		}
	}
	if len(keep) == 0 {
		return user
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !keep[field.Name] {
			f := v.Field(i)
			if f.CanSet() {
				zero := reflect.Zero(f.Type())
				f.Set(zero)
			}
		}
	}

	return user
}
