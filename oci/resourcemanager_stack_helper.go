// Copyright (c) 2017, 2020, Oracle and/or its affiliates. All rights reserved.
// Licensed under the Mozilla Public License v2.0

package oci

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	"github.com/oracle/oci-go-sdk/v26/common"
	"github.com/oracle/oci-go-sdk/v26/resourcemanager"
)

func createResourceManagerStack(resourceManagerClient resourcemanager.ResourceManagerClient, stackDisplayName string, compartment string) (string, error) {

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	f, err := zipWriter.Create("test.tf")
	if err != nil {
		return "", fmt.Errorf("[DEBUG] cannot create file for zip configuration: %v", err)
	}
	_, err = f.Write([]byte("provider oci{}"))
	if err != nil {
		return "", fmt.Errorf("[DEBUG] cannot write tf configuration to zip archive: %v", err)
	}
	err = zipWriter.Close()
	if err != nil {
		return "", fmt.Errorf("[DEBUG] cannot close zip writer: %v", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	// stack representation to assert in tests
	stackDescription := stackDisplayName + " Description"

	createStackResponse, err := resourceManagerClient.CreateStack(context.Background(), resourcemanager.CreateStackRequest{
		CreateStackDetails: resourcemanager.CreateStackDetails{
			CompartmentId: &compartment,
			ConfigSource: resourcemanager.CreateZipUploadConfigSourceDetails{
				ZipFileBase64Encoded: &encoded,
			},
			FreeformTags: map[string]string{"bar-key": "value"},
			DefinedTags:  map[string]map[string]interface{}{"example-tag-namespace-all": {"example-tag": "value"}},
			Description:  &stackDescription,
			DisplayName:  &stackDisplayName,
			Variables:    map[string]string{"var1": "value1", "var2": "value2", "var3": "value3"},
		},
		RequestMetadata: common.RequestMetadata{
			RetryPolicy: getRetryPolicy(false, "resourcemanager"),
		},
	})

	if err != nil {
		return "", fmt.Errorf("[DEBUG] cannot create ResourceManager stack: %v", err)
	}

	return *createStackResponse.Id, nil

}

func destroyResourceManagerStack(resourceManagerClient resourcemanager.ResourceManagerClient, resourceManagerStackId string) error {

	if resourceManagerStackId == "" {
		return fmt.Errorf("[DEBUG] ResourceManager StackId is not specified to delete the stack")
	}

	_, err := resourceManagerClient.DeleteStack(context.Background(), resourcemanager.DeleteStackRequest{
		StackId: &resourceManagerStackId,
		RequestMetadata: common.RequestMetadata{
			RetryPolicy: getRetryPolicy(true, "resourcemanager"),
		},
	})

	if err != nil {
		return fmt.Errorf("[DEBUG] failed to delete resource stack: %s with the error: %v", resourceManagerStackId, err)
	}

	return nil
}
