package flex

import (
	"encoding/xml"
	"errors"
	"fmt"
)

var (
	ErrInvalidToken        = errors.New("flex: token invalid or expired")
	ErrInvalidQueryID      = errors.New("flex: invalid query id")
	ErrStatementInProgress = errors.New("flex: statement still generating")
	ErrStatementTimeout    = errors.New("flex: statement generation timed out")
)

type FlexError struct {
	Code    string
	Message string
}

func (e *FlexError) Error() string {
	if e.Code == "" {
		return "flex: " + e.Message
	}
	return fmt.Sprintf("flex: %s (code %s)", e.Message, e.Code)
}

type flexStatementResponse struct {
	XMLName       xml.Name `xml:"FlexStatementResponse"`
	Status        string   `xml:"Status"`
	ReferenceCode string   `xml:"ReferenceCode"`
	URL           string   `xml:"Url"`
	ErrorCode     string   `xml:"ErrorCode"`
	ErrorMessage  string   `xml:"ErrorMessage"`
}

func parseStatementResponse(body []byte) (*flexStatementResponse, error) {
	var resp flexStatementResponse
	if err := xml.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("flex: decode FlexStatementResponse: %w", err)
	}
	return &resp, nil
}

func errorFromCode(code, msg string) error {
	switch code {
	case "1018":
		return ErrInvalidToken
	case "1019":
		return ErrStatementInProgress
	case "1020":
		return ErrInvalidQueryID
	}
	return &FlexError{Code: code, Message: msg}
}
