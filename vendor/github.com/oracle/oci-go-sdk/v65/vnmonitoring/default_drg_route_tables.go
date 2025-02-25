// Copyright (c) 2016, 2018, 2025, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Network Monitoring API
//
// Use the Network Monitoring API to troubleshoot routing and security issues for resources such as virtual cloud networks (VCNs) and compute instances. For more information, see the console
// documentation for the Network Path Analyzer (https://docs.oracle.com/iaas/Content/Network/Concepts/path_analyzer.htm) tool.
//

package vnmonitoring

import (
	"fmt"
	"github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// DefaultDrgRouteTables The default DRG route table for this DRG. Each network type
// has a default DRG route table.
// You can update a network type to use a different DRG route table, but
// each network type must have a default DRG route table. You cannot delete
// a default DRG route table.
type DefaultDrgRouteTables struct {

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the default DRG route table to be assigned to DRG attachments
	// of type VCN on creation.
	Vcn *string `mandatory:"false" json:"vcn"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the default DRG route table assigned to DRG attachments
	// of type IPSEC_TUNNEL on creation.
	IpsecTunnel *string `mandatory:"false" json:"ipsecTunnel"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the default DRG route table to be assigned to DRG attachments
	// of type VIRTUAL_CIRCUIT on creation.
	VirtualCircuit *string `mandatory:"false" json:"virtualCircuit"`

	// The OCID (https://docs.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the default DRG route table to be assigned to DRG attachments
	// of type REMOTE_PEERING_CONNECTION on creation.
	RemotePeeringConnection *string `mandatory:"false" json:"remotePeeringConnection"`
}

func (m DefaultDrgRouteTables) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m DefaultDrgRouteTables) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
