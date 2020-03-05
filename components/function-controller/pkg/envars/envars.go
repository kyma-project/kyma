package envars

import (
	"os"
	"strconv"
	"time"

	"github.com/kyma-project/kyma/components/function-controller/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
)

func BuildOptionParseEvnStrWithDefault(dst *string, key string, defaultval *string) func() error {
	return func() error {
		val, found := os.LookupEnv(key)
		if found {
			*dst = val
			return nil
		}
		if defaultval != nil {
			*dst = *defaultval
			return nil
		}
		return errors.NewEmptyValue(key)
	}
}

func BuildOptionParseEvnStr(dst *string, key string) func() error {
	return BuildOptionParseEvnStrWithDefault(dst, key, nil)
}

func BuildOptionParseEvnIntWithDefault(dst *int, key string, defaultval *int) func() error {
	return func() error {
		val, found := os.LookupEnv(key)
		if found {
			ival, err := strconv.Atoi(val)
			if err != nil {
				return err
			}
			*dst = ival
			return nil
		}
		if defaultval != nil {
			*dst = *defaultval
			return nil
		}
		return errors.NewEmptyValue(key)
	}
}

func BuildOptionParseEvnBool(dst *bool, key string) func() error {
	return func() error {
		val, found := os.LookupEnv(key)
		if !found {
			*dst = false
			return nil
		}
		bval, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		*dst = bval
		return nil
	}
}

func BuildOptionParseEvnQuantityWithDefault(dst *resource.Quantity, key string, defaultval *resource.Quantity) func() error {
	return func() error {
		val, found := os.LookupEnv(key)
		if found {
			quantity, err := resource.ParseQuantity(val)
			if err != nil {
				return errors.NewInvalidValue(err)
			}
			*dst = quantity
			return nil
		}
		if defaultval != nil {
			*dst = *defaultval
			return nil
		}
		return errors.NewEmptyValue(key)
	}
}

func BuildOptionParseEvnDurationWithDefault(dst *time.Duration, key string, defaultval *time.Duration) func() error {
	return func() error {
		val, found := os.LookupEnv(key)
		if found {
			tDuration, err := time.ParseDuration(val)
			if err != nil {
				return err
			}
			*dst = tDuration
			return nil
		}
		if defaultval != nil {
			*dst = *defaultval
			return nil
		}
		return errors.NewEmptyValue(key)
	}
}
