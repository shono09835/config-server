package integration_test

import (
	"bytes"
	"io"
	"net/http"
	"os/exec"
	"strings"

	. "github.com/shono09835/config-server/integration/support"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Supported HTTP Methods", func() {
	var session *gexec.Session

	BeforeEach(func() {
		SetupDB()

		var err error
		cmd := exec.Command(pathToConfigServer, pathToConfigFile)
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		WaitForServerToStart()
	})

	AfterEach(func() {
		session.Kill()
		Eventually(session).Should(gexec.Exit())
	})

	Describe("GET", func() {
		Describe("Lookup by name", func() {
			It("errors when name has invalid characters", func() {
				resp, err := SendGetRequestByName("sm!urf/garg$amel/cat")

				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(400))

				body, _ := io.ReadAll(resp.Body)
				Expect(strings.TrimSpace(string(body))).To(Equal(`{"error":"Name must consist of alphanumeric, underscores, dashes, and forward slashes"}`))
			})

			Context("when name does not exist in server", func() {
				It("responds with status 404", func() {
					resp, err := SendGetRequestByName("smurf")

					Expect(err).To(BeNil())
					Expect(resp.StatusCode).To(Equal(404))

					body, _ := io.ReadAll(resp.Body)
					Expect(strings.TrimSpace(string(body))).To(Equal(`{"error":"Name 'smurf' not found"}`))
				})
			})

			Context("when name exists in server", func() {
				It("responds with status 200", func() {
					_, err := SendPutRequest("smurf", "blue")
					Expect(err).To(BeNil())

					resp, err := SendGetRequestByName("smurf")

					Expect(err).To(BeNil())
					Expect(resp.StatusCode).To(Equal(200))
				})

				It("sends back id, name and value as json", func() {
					SendPutRequest("smurf", "blue") //nolint:errcheck

					resp, err := SendGetRequestByName("smurf")

					Expect(err).To(BeNil())

					resultMap := UnmarshalJSONString(resp.Body)

					data := resultMap["data"].([]interface{})
					entry := data[0].(map[string]interface{})

					Expect(entry["id"]).ToNot(BeNil())
					Expect(entry["name"]).To(Equal("smurf"))
					Expect(entry["value"]).To(Equal("blue"))
				})

				It("sends back ALL values sorted by ID", func() {
					SendPutRequest("smurf", "red")   //nolint:errcheck
					SendPutRequest("smurf", "green") //nolint:errcheck
					SendPutRequest("smurf", "blue")  //nolint:errcheck

					resp, err := SendGetRequestByName("smurf")

					Expect(err).To(BeNil())

					resultMap := UnmarshalJSONString(resp.Body)

					data := resultMap["data"].([]interface{})
					entry1 := data[0].(map[string]interface{})
					entry2 := data[1].(map[string]interface{})
					entry3 := data[2].(map[string]interface{})

					Expect(entry1["name"]).To(Equal("smurf"))
					Expect(entry1["value"]).To(Equal("blue"))

					Expect(entry2["name"]).To(Equal("smurf"))
					Expect(entry2["value"]).To(Equal("green"))

					Expect(entry3["name"]).To(Equal("smurf"))
					Expect(entry3["value"]).To(Equal("red"))
				})

				It("handles names with forward slashes", func() {
					name := "smurf/gar_gamel/c-at"

					SendPutRequest(name, "vroom") //nolint:errcheck

					resp, err := SendGetRequestByName(name)

					Expect(err).To(BeNil())

					resultMap := UnmarshalJSONString(resp.Body)

					data := resultMap["data"].([]interface{})
					entry := data[0].(map[string]interface{})

					Expect(entry["name"]).To(Equal(name))
					Expect(entry["value"]).To(Equal("vroom"))
				})
			})
		})

		Describe("Lookup by ID", func() {
			Context("when id does not exist in server", func() {
				It("responds with status 404", func() {
					resp, err := SendGetRequestByID("123")

					Expect(err).To(BeNil())
					Expect(resp.StatusCode).To(Equal(404))

					body, _ := io.ReadAll(resp.Body)
					Expect(strings.TrimSpace(string(body))).To(Equal(`{"error":"ID '123' not found"}`))
				})
			})

			Context("when id exists in server", func() {
				It("responds with status 200", func() {
					putResponse, _ := SendPutRequest("smurf", "blue")
					config := UnmarshalJSONString(putResponse.Body)
					id := config["id"].(string)

					resp, err := SendGetRequestByID(id)

					Expect(err).To(BeNil())
					Expect(resp.StatusCode).To(Equal(200))
				})

				It("sends back value along with name as json", func() {
					putResponse, _ := SendPutRequest("annie", "diane")
					config := UnmarshalJSONString(putResponse.Body)
					id := config["id"].(string)

					resp, err := SendGetRequestByID(id)

					Expect(err).To(BeNil())

					resultMap := UnmarshalJSONString(resp.Body)

					Expect(resultMap["name"]).To(Equal("annie"))
					Expect(resultMap["value"]).To(Equal("diane"))
					Expect(resultMap["id"]).To(Equal(id))
				})
			})
		})
	})

	Describe("PUT", func() {
		It("fails if content-type in the header is not set to application/json", func() {
			requestBytes := bytes.NewReader([]byte(`{"name":"blah", "value":"smurf"`))
			req, _ := http.NewRequest("PUT", ServerURL+"/v1/data/", requestBytes)
			req.Header.Add("Authorization", "bearer "+ValidToken())

			resp, err := HTTPSClient.Do(req)
			Expect(resp.StatusCode).To(Equal(415))
			Expect(err).To(BeNil())

			body, _ := io.ReadAll(resp.Body)
			Expect(strings.TrimSpace(string(body))).To(Equal(`{"error":"Unsupported Media Type - Accepts application/json only"}`))
		})

		It("errors when name has invalid characters", func() {
			resp, err := SendPutRequest("sm!urf/garg$amel/cat", "value")

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(400))

			body, _ := io.ReadAll(resp.Body)
			Expect(strings.TrimSpace(string(body))).To(Equal(`{"error":"Name must consist of alphanumeric, underscores, dashes, and forward slashes"}`))
		})

		Context("when name does NOT exist in server", func() {
			It("responds with value & id", func() {
				resp, err := SendPutRequest("cross", "fit")

				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(200))

				resultMap := UnmarshalJSONString(resp.Body)
				Expect(resultMap["name"]).To(Equal("cross"))
				Expect(resultMap["value"]).To(Equal("fit"))
			})

			It("responds with status 200 when value is successfully stored", func() {
				resp, err := SendPutRequest("smurf", "blue")

				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(200))
			})

			Context("when value is empty string", func() {
				It("is stored and responds with value & id", func() {
					resp, err := SendPutRequest("crossfit", "")

					Expect(err).To(BeNil())
					Expect(resp.StatusCode).To(Equal(200))

					resultMap := UnmarshalJSONString(resp.Body)
					Expect(resultMap["name"]).To(Equal("crossfit"))
					Expect(resultMap["value"]).To(Equal(""))
				})
			})
			Context("when value is nil", func() {
				It("is stored and responds with value & id", func() {
					resp, err := SendPutRequest("crossfit", nil)

					Expect(err).To(BeNil())
					Expect(resp.StatusCode).To(Equal(200))

					resultMap := UnmarshalJSONString(resp.Body)
					Expect(resultMap["name"]).To(Equal("crossfit"))
					Expect(resultMap["value"]).To(BeNil())
				})
			})
		})

		Context("when name exists in server", func() {
			It("updates the value", func() {
				SendPutRequest("smurf", "blue") //nolint:errcheck

				getResp, _ := SendGetRequestByName("smurf")

				resultMap := UnmarshalJSONString(getResp.Body)
				data := resultMap["data"].([]interface{})
				entry := data[0].(map[string]interface{})

				Expect(entry["name"]).To(Equal("smurf"))
				Expect(entry["value"]).To(Equal("blue"))

				SendPutRequest("smurf", "red") //nolint:errcheck
				getResp, _ = SendGetRequestByName("smurf")

				resultMap = UnmarshalJSONString(getResp.Body)
				data = resultMap["data"].([]interface{})
				entry = data[0].(map[string]interface{})

				Expect(entry["name"]).To(Equal("smurf"))
				Expect(entry["value"]).To(Equal("red"))
			})
		})
	})

	Describe("POST", func() {
		It("fails if content-type in the header is not set to application/json", func() {
			requestBytes := bytes.NewReader([]byte(`{"name":"blah", "type":"password","parameters":{}}`))
			req, _ := http.NewRequest("POST", ServerURL+"/v1/data/", requestBytes)
			req.Header.Add("Authorization", "bearer "+ValidToken())

			resp, err := HTTPSClient.Do(req)
			Expect(resp.StatusCode).To(Equal(415))
			Expect(err).To(BeNil())

			body, _ := io.ReadAll(resp.Body)
			Expect(strings.TrimSpace(string(body))).To(Equal(`{"error":"Unsupported Media Type - Accepts application/json only"}`))
		})

		It("fails if is_ca is set but ca is NOT", func() {
			response, err := SendPostRequest("certificate-name", "self-signed-certificate", "", false)

			Expect(response.StatusCode).To(Equal(400))
			Expect(err).To(BeNil())

			body, _ := io.ReadAll(response.Body)

			Expect(strings.TrimSpace(string(body))).To(Equal(`{"error":"Missing required CA name"}`))
		})

		It("generates a new id and password for a new name", func() {
			resp, _ := SendPostRequest("password-name", "password", "", false)
			result := UnmarshalJSONString(resp.Body)

			Expect(result["id"]).ToNot(BeNil())
			Expect(result["name"]).To(Equal("password-name"))
			Expect(result["value"]).To(MatchRegexp("[a-z0-9]{20}"))
		})

		It("generates a new id and certificate for a new name", func() {
			SendPostRequest("my-ca", "root-certificate-ca", "", false) //nolint:errcheck

			resp, _ := SendPostRequest("some-signed-certificate-name", "certificate", "", false)

			result := UnmarshalJSONString(resp.Body)

			Expect(result["id"]).ToNot(BeNil())
			Expect(result["name"]).To(Equal("some-signed-certificate-name"))

			value := result["value"].(map[string]interface{})
			cert, _ := ParseCertString(value["certificate"].(string))

			Expect(cert.DNSNames).Should(ContainElement("signed-an1"))
			Expect(cert.DNSNames).Should(ContainElement("signed-an1"))
			Expect(cert.Subject.CommonName).To(Equal("some-signed-cn1"))

			Expect(cert.IsCA).To(BeFalse())

			Expect(cert.Issuer.Organization).To(ContainElement("Cloud Foundry"))
			Expect(cert.Issuer.Country).To(ContainElement("USA"))
			Expect(cert.Issuer.CommonName).To(Equal("some-root-certificate-ca-cn1"))
		})

		It("generates a new id and root ca certificate for a new name", func() {
			resp, _ := SendPostRequest("some-root-certificate-name", "root-certificate-ca", "", false)
			result := UnmarshalJSONString(resp.Body)

			Expect(result["id"]).ToNot(BeNil())
			Expect(result["name"]).To(Equal("some-root-certificate-name"))

			value := result["value"].(map[string]interface{})

			cert, _ := ParseCertString(value["certificate"].(string))

			Expect(cert.DNSNames).Should(BeEmpty())
			Expect(cert.IPAddresses).Should(BeEmpty())
			Expect(cert.IsCA).Should(BeTrue())
			Expect(cert.Subject.CommonName).To(Equal("some-root-certificate-ca-cn1"))

			Expect(cert.Issuer.Organization).To(ContainElement("Cloud Foundry"))
			Expect(cert.Issuer.Country).To(ContainElement("USA"))
		})

		It("generates a new id and intermediate ca certificate for a new name", func() {
			SendPostRequest("my-ca", "root-certificate-ca", "", false) //nolint:errcheck

			resp, _ := SendPostRequest("certificate-name", "intermediate-certificate-ca", "", false)
			result := UnmarshalJSONString(resp.Body)

			Expect(result["id"]).ToNot(BeNil())
			Expect(result["name"]).To(Equal("certificate-name"))

			value := result["value"].(map[string]interface{})
			cert, _ := ParseCertString(value["certificate"].(string))

			Expect(cert.DNSNames).Should(BeEmpty())
			Expect(cert.IPAddresses).Should(BeEmpty())
			Expect(cert.IsCA).Should(BeTrue())
			Expect(cert.Subject.CommonName).To(Equal("some-intermediate-certificate-ca-cn1"))

			Expect(cert.Issuer.Organization).To(ContainElement("Cloud Foundry"))
			Expect(cert.Issuer.Country).To(ContainElement("USA"))
			Expect(cert.Issuer.CommonName).To(Equal("some-root-certificate-ca-cn1"))
		})

		It("generates certificates with unique SKI and same AKI when sharing CA", func() {
			SendPostRequest("my-ca", "root-certificate-ca", "", false) //nolint:errcheck

			firstCertResp, _ := SendPostRequest("first-certificate-name", "certificate", "", false)
			firstCertResult := UnmarshalJSONString(firstCertResp.Body)
			firstCertValue := firstCertResult["value"].(map[string]interface{})
			firstCert, _ := ParseCertString(firstCertValue["certificate"].(string))

			secondCertResp, _ := SendPostRequest("second-certificate-name", "certificate", "", false)
			secondCertResult := UnmarshalJSONString(secondCertResp.Body)
			secondCertValue := secondCertResult["value"].(map[string]interface{})
			secondCert, _ := ParseCertString(secondCertValue["certificate"].(string))

			Expect(firstCert).ToNot(Equal(secondCert))
			Expect(firstCert.SubjectKeyId).ToNot(Equal(secondCert.SubjectKeyId))
			Expect(firstCert.AuthorityKeyId).ToNot(BeNil())
			Expect(firstCert.AuthorityKeyId).To(Equal(secondCert.AuthorityKeyId))
		})

		Context("when mode is set to converge", func() {
			It("generates a new id and variable if the parameters are the different", func() {
				resp, _ := SendPostRequest("password-name", "password", "", true)
				result := UnmarshalJSONString(resp.Body)

				Expect(len(result["value"].(string))).To(Equal(20))

				resp2, _ := SendPostRequest("password-name", "password", "\"length\": 8", true)
				result2 := UnmarshalJSONString(resp2.Body)

				Expect(len(result2["value"].(string))).To(Equal(8))

				Expect(result2["id"]).ToNot(BeNil())
				Expect(result2["name"]).To(Equal("password-name"))
				Expect(result2["value"]).ToNot(Equal(result["value"]))
			})

			It("generates a new id and variable if the parameters are the same", func() {
				resp, _ := SendPostRequest("password-name", "password", "\"length\": 8", true)
				result := UnmarshalJSONString(resp.Body)
				resp2, _ := SendPostRequest("password-name", "password", "\"length\": 8", true)
				result2 := UnmarshalJSONString(resp2.Body)

				Expect(result2["id"]).ToNot(BeNil())
				Expect(result2["name"]).To(Equal("password-name"))
				Expect(result2["value"]).To(Equal(result["value"]))
			})
		})
	})

	Describe("DELETE", func() {

		It("deletes ALL entries for for the name", func() {
			SendPutRequest("smurf", "green") //nolint:errcheck
			SendPutRequest("smurf", "blue")  //nolint:errcheck

			resp, err := SendGetRequestByName("smurf")
			Expect(err).NotTo(HaveOccurred())
			resultMap := UnmarshalJSONString(resp.Body)

			data := resultMap["data"].([]interface{})
			Expect(len(data)).To(Equal(2))

			SendDeleteRequest("smurf") //nolint:errcheck

			resp, err = SendGetRequestByName("smurf")
			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))

			body, _ := io.ReadAll(resp.Body)
			Expect(strings.TrimSpace(string(body))).To(Equal(`{"error":"Name 'smurf' not found"}`))
		})

		It("returns 204 No Content when deletion is successful", func() {
			SendPutRequest("smurf", "blue") //nolint:errcheck

			resp, err := SendDeleteRequest("smurf")

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})

		It("returns 404 Not found when configuration with name does not exist", func() {
			resp, err := SendDeleteRequest("smurf")

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))

			body, _ := io.ReadAll(resp.Body)
			Expect(strings.TrimSpace(string(body))).To(Equal(`{"error":"Name 'smurf' not found"}`))
		})
	})
})
