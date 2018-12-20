package validation_test

import (
	. "autoscaler/api/policymanager/validation"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PolicyValidator", func() {
	var (
		policyValidator *PolicyValidator
	)
	BeforeEach(func() {
		policyValidator = NewPolicyValidator("/Users/i332436/workspace/opensource/app-autoscaler/src/autoscaler/api/policymanager/policy_json.schema.json")
	})

	Context("If the policy schema is valid", func() {

		It("should fail to validate", func() {
			err := policyValidator.ValidatePolicy(`{
				"instance_min_count":1,
				"instance_max_count":3,
				"scaling_rules":[
					{
						"metric_type":"memoryused",
						"breach_duration_secs":600,
						"threshold":30,
						"operator":"<",
						"cool_down_secs":300,
						"adjustment":"-1"
					}
				]
			}`)
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("If the policy schema is invalid", func() {

		It("should fail to validate", func() {
			err := policyValidator.ValidatePolicy(`{
				"instance_min_count":1,
				"scaling_rules":[
					{
						"metric_type":"memoryused",
						"breach_duration_secs":600,
						"threshold":30,
						"operator":5,
						"cool_down_secs":300,
						"adjustment":"-1"
					}
				]
			}`)
			fmt.Println(err)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("instance_min_count is greater than instance_max_count", func() {
		It("It should fail to validate", func() {
			err := policyValidator.ValidatePolicy(`{
				"instance_min_count":5,
				"instance_max_count":2,
				"scaling_rules":[
					{
						"metric_type":"memoryused",
						"breach_duration_secs":600,
						"threshold":30,
						"operator":"<",
						"cool_down_secs":300,
						"adjustment":"-1"
					}
				]
			}`)
			fmt.Println(err)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("If specific date schedules are overlapping", func() {
		It("It should fail to validate", func() {
			err := policyValidator.ValidatePolicy(`{
				"instance_min_count":1,
				"instance_max_count":2,
				"scaling_rules":[
					{
						"metric_type":"memoryused",
						"breach_duration_secs":600,
						"threshold":30,
						"operator":"<",
						"cool_down_secs":300,
						"adjustment":"-1"
					}
				],
				"schedules":{
					"timezone":"Asia/Shanghai",
					"specific_date":[
						{
							"start_date_time":"2020-01-02T10:00",
							"end_date_time":"2020-06-15T13:59",
							"instance_min_count":10,
							"instance_max_count":4,
							"initial_min_instance_count":2
						},
						{
							"start_date_time":"2020-01-04T20:00",
							"end_date_time":"2020-02-19T23:15",
							"instance_min_count":2,
							"instance_max_count":5,
							"initial_min_instance_count":3
						}
					]
				}
			}`)
			fmt.Println(err)
			Expect(err).To(HaveOccurred())
		})
	})
})
