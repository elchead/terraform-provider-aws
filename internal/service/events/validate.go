// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"fmt"
	"regexp"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func validateRuleName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 64 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 64 characters: %q", k, value))
	}

	// http://docs.aws.amazon.com/eventbridge/latest/APIReference/API_PutRule.html
	pattern := `^[\.\-_A-Za-z0-9]+$`
	if !regexache.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q doesn't comply with restrictions (%q): %q",
			k, pattern, value))
	}

	return
}

func validateTargetID(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 64 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 64 characters: %q", k, value))
	}

	// http://docs.aws.amazon.com/eventbridge/latest/APIReference/API_Target.html
	pattern := `^[\.\-_A-Za-z0-9]+$`
	if !regexache.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q doesn't comply with restrictions (%q): %q",
			k, pattern, value))
	}

	return
}

func mapKeysDoNotMatch(r *regexp.Regexp, message string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		m, ok := i.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be map", k))
			return warnings, errors
		}

		for key := range m {
			if ok := r.MatchString(key); ok {
				errors = append(errors, fmt.Errorf("%s: %s: %s", k, message, key))
			}
		}

		return warnings, errors
	}
}

func mapMaxItems(max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		m, ok := i.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be map", k))
			return warnings, errors
		}

		if len(m) > max {
			errors = append(errors, fmt.Errorf("expected number of items in %s to be less than or equal to %d, got %d", k, max, len(m)))
		}

		return warnings, errors
	}
}

var validArchiveName = validation.All(
	validation.StringLenBetween(1, 48),
	validation.StringMatch(regexache.MustCompile(`^`+validNameCharClass+`$`), ""),
)

var validBusName = validation.All(
	validation.StringLenBetween(1, 256),
	validBusNameFormat,
)

var validBusNameOrARN = validation.All(
	validation.StringLenBetween(1, 1600),
	validation.Any(
		verify.ValidARNCheck(eventBusARNCheck),
		validBusNameFormat,
	),
)

var validBusNameFormat = validation.StringMatch(regexache.MustCompile(`^`+validBusNameCharClass+`$`), "")

func eventBusARNCheck(v any, k string, arn arn.ARN) (ws []string, errors []error) {
	if !isEventBusARN(arn) {
		errors = append(errors, fmt.Errorf("%q (%s) is not a valid Event Bus ARN", k, v))
	}
	return
}

var validSourceName = validation.All(
	validation.StringLenBetween(1, 256),
	validation.StringMatch(regexache.MustCompile(`^aws\.partner(/`+validNameCharClass+`){2,}$`), ""),
)

var validCustomEventBusName = validation.All(
	validation.StringLenBetween(1, 256),
	validation.StringDoesNotMatch(regexache.MustCompile(`^default$`), "cannot be 'default'"),
)

func isEventBusARN(arn arn.ARN) bool {
	if arn.Service != eventbridge.EndpointsID {
		return false
	}

	re := regexache.MustCompile(`^event-bus/(` + validBusNameCharClass + `)$`)

	return re.MatchString(arn.Resource)
}

var validNameChars = `\.\-_A-Za-z0-9`
var validNameCharClass = `[` + validNameChars + `]+`

var validBusNameCharPattern = `/` + validNameChars
var validBusNameCharClass = `[` + validBusNameCharPattern + `]+`
