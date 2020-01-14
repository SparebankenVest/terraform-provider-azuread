package azuread

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/ar"
	"github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/graph"
)

func testCheckADApplicationCertificateExists(name string) resource.TestCheckFunc { //nolint unparam
	return nil

	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ArmClient).applicationsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %q", name)
		}

		id, err := graph.ParsePasswordCredentialId(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing Application Password Credential ID: %v", err)
		}
		resp, err := client.Get(ctx, id.ObjectId)
		if err != nil {
			if ar.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Azure AD Application  %q does not exist", id.ObjectId)
			}
			return fmt.Errorf("Bad: Get on Azure AD applicationsClient: %+v", err)
		}

		credentials, err := client.ListPasswordCredentials(ctx, id.ObjectId)
		if err != nil {
			return fmt.Errorf("Error Listing Password Credentials for Application %q: %+v", id.ObjectId, err)
		}

		cred := graph.PasswordCredentialResultFindByKeyId(credentials, id.KeyId)
		if cred != nil {
			return nil
		}

		return fmt.Errorf("Password Credential %q was not found in Application %q", id.KeyId, id.ObjectId)
	}
}

func TestAccAzureADApplicationCertificate_basic(t *testing.T) {
	resourceName := "azuread_application_certificate.test"
	applicationId := uuid.New().String()
	value := uuid.New().String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckADApplicationCertificateCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccADObjectPasswordApplication_basic(applicationId, value),
				Check: resource.ComposeTestCheckFunc(
					// can't assert on Value since it's not returned
					testCheckADApplicationPasswordExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "start_date"),
					resource.TestCheckResourceAttrSet(resourceName, "key_id"),
					resource.TestCheckResourceAttr(resourceName, "end_date", "2099-01-01T01:02:03Z"),
				),
			},
		},
	})
}

func testCheckADApplicationCertificateCheckDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		client := testAccProvider.Meta().(*ArmClient).applicationsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		if rs.Type != "azuread_service_principal_certificate" {
			continue
		}

		id, err := graph.ParsePasswordCredentialId(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing Service Principal Password Credential ID: %v", err)
		}

		resp, err := client.Get(ctx, id.ObjectId)
		if err != nil {
			if ar.ResponseWasNotFound(resp.Response) {
				return nil
			}

			return err
		}

		return fmt.Errorf("Azure AD Service Principal Password Credential still exists:\n%#v", resp)
	}

	return nil
}

func testAccADApplicationCertificate_template(applicationId string) string {
	return fmt.Sprintf(`
resource "azuread_application" "test" {
  name = "acctestApp-%s"
}
`, applicationId)
}

func testAccADApplicationCertificate_basic(applicationId, value string) string {
	return fmt.Sprintf(`
%s

resource "azuread_application_certificate" "test" {
  application_object_id = "${azuread_application.test.id}"
  certificate = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZXVENDQTBHZ0F3SUJBZ0lVZnBDUWVuNmhWMGpKa3RlSzJ4RGZETkFRVlNFd0RRWUpLb1pJaHZjTkFRRUwKQlFBd1BERUxNQWtHQTFVRUJoTUNUazh4RXpBUkJnTlZCQWdNQ2xOdmJXVXRVM1JoZEdVeEdEQVdCZ05WQkFvTQpEMU5sYkdZdFUybG5ibVZrSUV4MFpEQWVGdzB5TURBeE1UUXhNVE0yTVRkYUZ3MHpNREF4TVRFeE1UTTJNVGRhCk1Ed3hDekFKQmdOVkJBWVRBazVQTVJNd0VRWURWUVFJREFwVGIyMWxMVk4wWVhSbE1SZ3dGZ1lEVlFRS0RBOVQKWld4bUxWTnBaMjVsWkNCTWRHUXdnZ0lpTUEwR0NTcUdTSWIzRFFFQkFRVUFBNElDRHdBd2dnSUtBb0lDQVFDcQpsckpGV09qL0pUOTRyb0FhS2FzRDFTU3grTjlrTU5XOWtsWTUrMGhCZXJHbm05WE1pU241cnBIUGYzcitqUEpiCm1XL05DZFdFeldmN1FCNlhrT3ZVMVRmcC9Wc05vNEVtSlM3NkxoUTF5cTFQcEhqMGRvSExnY3RJcmRwUVdpeW0KTVlMMU11VEVERFkxM3ROOEJmRlY2cG80RWRCV1pOTFlQTXlxQUR5TlFuVHllVEIvSFNObTlQaHR2WDlsUk9PNApwUDBuM1hReG1FYTJqekw1Tm16dzZtQ2pFMFhlOHZicjJqVnI4QVk0ZW1qM3BJZjlDVHE1ZGVIQzBqTzVyUmVwCituYSs3WlFaNVFJaGlVWE96YW53V25pZXdUN0Erdi9GY1Bsc2JXNHVaSXRqYkQvL3lqYnc3cmc5ZFpWU3RpUlMKa0JtcENUZEUxVWt3WW9OVm5CV0gvQktoMERqMGRRc01hcG5UVU02OHlUR3hiOGtTTDM5ZngzVGN1ZzZ3MHBXbApYZEdaVERHK2cyRk1COHN5a3pTVWVuT2VGeWdzb2cvaEVOaGNpVzhyK3pPVEtlVlh0R0ppMVZnZnZQTWQ5aHVJClkzeXhINVN2dTBSTTJNV2pueW44Y202NEJwKzZSMEpYZlpqbWVhUFVneWdqdW0xNUVKVXEvRTJaNFlLb3gxQW4KbTVSSGRldzhpbS9FRnJPSDVNSm95VUtybUNORUZ1ZFhoM29JVW90TnkraVd3OXFzMUhZZDVlYWtuS0NkZGJZLwpjZzNIQ0FtSmhUaTlpS2hWcDc0Ump6WWJ2STNrY2o3SXN2Yk42RGtHNEE0NXNNNWhCUUJRbWdxblRwTkxYYkdKCndiaWxmMDlxRlZWdldCS2o2cDJmV1ZTQW5lcnVTS2RvQXR6R0FnRXJBd0lEQVFBQm8xTXdVVEFkQmdOVkhRNEUKRmdRVWM3K1ZVNWlnaXNVcDJ2dE9HaXJXUFd5RDYrQXdId1lEVlIwakJCZ3dGb0FVYzcrVlU1aWdpc1VwMnZ0TwpHaXJXUFd5RDYrQXdEd1lEVlIwVEFRSC9CQVV3QXdFQi96QU5CZ2txaGtpRzl3MEJBUXNGQUFPQ0FnRUFqNGpYCnlzS3V6MHBWQVhxcjAwYlFYcnRKRzl4Y0dJWWpJNmlacmhGMzlKeFlZR3lWZXNOb0NkR2ZwVkcxZVNPRjJYWHkKdHZaV0I3WTZBZmY2cWVYeDVvd1FtYkx3Wm9zS21TNG16UnlpWVhSWFliTnQ5am5zclNYbGE1VVpEL1pDVHBGUQoyZjJRYm91MnFsN3RFc01neDRnclM1SEQvVHdES0c0MnN0cWppTGljUG1aOFlONUtCWm04TXhEbEtkZmlXS2VqCmtzbHNFRjU1MzBHL1dsckU4ZUVQc29ZUGF2QW1wc2tPQVJiRDEyTHJVQWs3NDRNeWUrMTdwK0pSOFExbTgyQzgKQU9vaHlFdHdLSi9VVEo4Zmhpa2pUaDVoK3djWlZ5ZDlSdkRRWnI5dGlWYlZkK0JKOGg2cVQwRCtIdk5OZU9udgpLamZjZU5WckRxck1yRW81aWttSnhlWER0R0tXSTVvZkpBYjhJN0htTXdIemVSQzR5ZHVqbTBkbHBxMndpR2ZGClI3Mnh3eXorNmNXeXFEMHhoS2ZjcGJBNVA3THdrQVkxR0M5cThzV0VEUnJYMFhibWRwdDV6ZVY3Z24xVGpHNE0KM2w1TkJJYThjZlAwU0VIUlJwSFpraWFQUDJTMEI2eEpiRHBGajZkMUgwQWlXTHBTS1VobElBUzhtbXE3Q1RsUQorZktqMWlwQk1YVnRSY0xGa0JkYWk1dmhOSDYvR1R3ZXdCcEJpekZvcHlEMVRNMVJncURNZHdmN2ZEelUyekduClNuQW1jVnpMd2h5UFd1ZHVlZzZtWTMxRTYySVJQNmZKSUp5b2Q0L3NRN0JPSG9EaVdyQ2h5dVdOM3RtQzdXeWMKTXJidWxiSXQyS2pERTJSeUU3RUJrejhMM3c2NlVBdjVGWEtaZVpBPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
}
`, testAccADApplicationCertificate_template(applicationId), value)
}
