package validation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPolicyvalidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PolicyManager Suite")
}
