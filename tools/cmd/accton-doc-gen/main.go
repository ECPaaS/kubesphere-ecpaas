package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"

	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/informers"
	virtualizationv1 "kubesphere.io/kubesphere/pkg/kapis/virtualization/v1"
	volumev1alpha1 "kubesphere.io/kubesphere/pkg/kapis/volume/v1alpha1"
	vpcv1 "kubesphere.io/kubesphere/pkg/kapis/vpc/v1"
	"kubesphere.io/kubesphere/pkg/version"

	"k8s.io/klog"
)

var output string

func init() {
	flag.StringVar(&output, "output", "./api/ac-openapi-spec/swagger.json", "--output=./api.json")
}

func main() {

	flag.Parse()
	swaggerSpec := generateSwaggerJson()

	err := validateSpec(swaggerSpec)
	if err != nil {
		klog.Warningf("Swagger specification has errors")
	}
}

func validateSpec(apiSpec []byte) error {

	swaggerDoc, err := loads.Analyzed(apiSpec, "")
	if err != nil {
		return err
	}

	// Attempts to report about all errors
	validate.SetContinueOnErrors(true)

	v := validate.NewSpecValidator(swaggerDoc.Schema(), strfmt.Default)
	result, _ := v.Validate(swaggerDoc)

	if result.HasWarnings() {
		log.Printf("See warnings below:\n")
		for _, desc := range result.Warnings {
			log.Printf("- WARNING: %s\n", desc.Error())
		}

	}
	if result.HasErrors() {
		str := fmt.Sprintf("The swagger spec is invalid against swagger specification %s.\nSee errors below:\n", swaggerDoc.Version())
		for _, desc := range result.Errors {
			str += fmt.Sprintf("- %s\n", desc.Error())
		}
		log.Println(str)
		return errors.New(str)
	}

	return nil
}

func generateSwaggerJson() []byte {

	container := runtime.Container
	informerFactory := informers.NewNullInformerFactory()

	urlruntime.Must(vpcv1.AddToContainer(container, informerFactory, nil, nil))
	urlruntime.Must(volumev1alpha1.AddToContainer(container, nil, nil, nil))
	urlruntime.Must(virtualizationv1.AddToContainer(container, nil, nil, nil, informerFactory))

	config := restfulspec.Config{
		WebServices:                   container.RegisteredWebServices(),
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}

	swagger := restfulspec.BuildSwagger(config)

	data, _ := json.MarshalIndent(swagger, "", "  ")
	err := os.WriteFile(output, data, 420)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("successfully written to %s", output)

	return data
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "KubeSphere",
			Description: "KubeSphere OpenAPI",
			Version:     version.Get().GitVersion,
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "KubeSphere",
					URL:   "https://kubesphere.io/",
					Email: "kubesphere@yunify.com",
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "Apache 2.0",
					URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
				},
			},
		},
	}

	// setup security definitions
	swo.SecurityDefinitions = map[string]*spec.SecurityScheme{
		"jwt": spec.APIKeyAuth("Authorization", "header"),
	}
	swo.Security = []map[string][]string{{"jwt": []string{}}}
}
