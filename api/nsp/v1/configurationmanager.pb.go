//
//Copyright (c) 2021-2022 Nordix Foundation
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.19.1
// source: api/nsp/v1/configurationmanager.proto

package v1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_api_nsp_v1_configurationmanager_proto protoreflect.FileDescriptor

var file_api_nsp_v1_configurationmanager_proto_rawDesc = []byte{
	0x0a, 0x25, 0x61, 0x70, 0x69, 0x2f, 0x6e, 0x73, 0x70, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x6e, 0x73, 0x70, 0x2e, 0x76, 0x31, 0x1a,
	0x16, 0x61, 0x70, 0x69, 0x2f, 0x6e, 0x73, 0x70, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xb3, 0x03, 0x0a, 0x14, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72,
	0x12, 0x39, 0x0a, 0x0b, 0x57, 0x61, 0x74, 0x63, 0x68, 0x54, 0x72, 0x65, 0x6e, 0x63, 0x68, 0x12,
	0x0e, 0x2e, 0x6e, 0x73, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x72, 0x65, 0x6e, 0x63, 0x68, 0x1a,
	0x16, 0x2e, 0x6e, 0x73, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x72, 0x65, 0x6e, 0x63, 0x68, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x3c, 0x0a, 0x0c, 0x57,
	0x61, 0x74, 0x63, 0x68, 0x43, 0x6f, 0x6e, 0x64, 0x75, 0x69, 0x74, 0x12, 0x0f, 0x2e, 0x6e, 0x73,
	0x70, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x64, 0x75, 0x69, 0x74, 0x1a, 0x17, 0x2e, 0x6e,
	0x73, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x64, 0x75, 0x69, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x39, 0x0a, 0x0b, 0x57, 0x61, 0x74,
	0x63, 0x68, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x0e, 0x2e, 0x6e, 0x73, 0x70, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x1a, 0x16, 0x2e, 0x6e, 0x73, 0x70, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x30, 0x01, 0x12, 0x33, 0x0a, 0x09, 0x57, 0x61, 0x74, 0x63, 0x68, 0x46, 0x6c, 0x6f,
	0x77, 0x12, 0x0c, 0x2e, 0x6e, 0x73, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x6c, 0x6f, 0x77, 0x1a,
	0x14, 0x2e, 0x6e, 0x73, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x6c, 0x6f, 0x77, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x30, 0x0a, 0x08, 0x57, 0x61, 0x74,
	0x63, 0x68, 0x56, 0x69, 0x70, 0x12, 0x0b, 0x2e, 0x6e, 0x73, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x56,
	0x69, 0x70, 0x1a, 0x13, 0x2e, 0x6e, 0x73, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x69, 0x70, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x42, 0x0a, 0x0e, 0x57,
	0x61, 0x74, 0x63, 0x68, 0x41, 0x74, 0x74, 0x72, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x11, 0x2e,
	0x6e, 0x73, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x74, 0x74, 0x72, 0x61, 0x63, 0x74, 0x6f, 0x72,
	0x1a, 0x19, 0x2e, 0x6e, 0x73, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x74, 0x74, 0x72, 0x61, 0x63,
	0x74, 0x6f, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12,
	0x3c, 0x0a, 0x0c, 0x57, 0x61, 0x74, 0x63, 0x68, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x12,
	0x0f, 0x2e, 0x6e, 0x73, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79,
	0x1a, 0x17, 0x2e, 0x6e, 0x73, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61,
	0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x42, 0x26, 0x5a,
	0x24, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6e, 0x6f, 0x72, 0x64,
	0x69, 0x78, 0x2f, 0x6d, 0x65, 0x72, 0x69, 0x64, 0x69, 0x6f, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6e,
	0x73, 0x70, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_api_nsp_v1_configurationmanager_proto_goTypes = []interface{}{
	(*Trench)(nil),            // 0: nsp.v1.Trench
	(*Conduit)(nil),           // 1: nsp.v1.Conduit
	(*Stream)(nil),            // 2: nsp.v1.Stream
	(*Flow)(nil),              // 3: nsp.v1.Flow
	(*Vip)(nil),               // 4: nsp.v1.Vip
	(*Attractor)(nil),         // 5: nsp.v1.Attractor
	(*Gateway)(nil),           // 6: nsp.v1.Gateway
	(*TrenchResponse)(nil),    // 7: nsp.v1.TrenchResponse
	(*ConduitResponse)(nil),   // 8: nsp.v1.ConduitResponse
	(*StreamResponse)(nil),    // 9: nsp.v1.StreamResponse
	(*FlowResponse)(nil),      // 10: nsp.v1.FlowResponse
	(*VipResponse)(nil),       // 11: nsp.v1.VipResponse
	(*AttractorResponse)(nil), // 12: nsp.v1.AttractorResponse
	(*GatewayResponse)(nil),   // 13: nsp.v1.GatewayResponse
}
var file_api_nsp_v1_configurationmanager_proto_depIdxs = []int32{
	0,  // 0: nsp.v1.ConfigurationManager.WatchTrench:input_type -> nsp.v1.Trench
	1,  // 1: nsp.v1.ConfigurationManager.WatchConduit:input_type -> nsp.v1.Conduit
	2,  // 2: nsp.v1.ConfigurationManager.WatchStream:input_type -> nsp.v1.Stream
	3,  // 3: nsp.v1.ConfigurationManager.WatchFlow:input_type -> nsp.v1.Flow
	4,  // 4: nsp.v1.ConfigurationManager.WatchVip:input_type -> nsp.v1.Vip
	5,  // 5: nsp.v1.ConfigurationManager.WatchAttractor:input_type -> nsp.v1.Attractor
	6,  // 6: nsp.v1.ConfigurationManager.WatchGateway:input_type -> nsp.v1.Gateway
	7,  // 7: nsp.v1.ConfigurationManager.WatchTrench:output_type -> nsp.v1.TrenchResponse
	8,  // 8: nsp.v1.ConfigurationManager.WatchConduit:output_type -> nsp.v1.ConduitResponse
	9,  // 9: nsp.v1.ConfigurationManager.WatchStream:output_type -> nsp.v1.StreamResponse
	10, // 10: nsp.v1.ConfigurationManager.WatchFlow:output_type -> nsp.v1.FlowResponse
	11, // 11: nsp.v1.ConfigurationManager.WatchVip:output_type -> nsp.v1.VipResponse
	12, // 12: nsp.v1.ConfigurationManager.WatchAttractor:output_type -> nsp.v1.AttractorResponse
	13, // 13: nsp.v1.ConfigurationManager.WatchGateway:output_type -> nsp.v1.GatewayResponse
	7,  // [7:14] is the sub-list for method output_type
	0,  // [0:7] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_api_nsp_v1_configurationmanager_proto_init() }
func file_api_nsp_v1_configurationmanager_proto_init() {
	if File_api_nsp_v1_configurationmanager_proto != nil {
		return
	}
	file_api_nsp_v1_model_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_nsp_v1_configurationmanager_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_nsp_v1_configurationmanager_proto_goTypes,
		DependencyIndexes: file_api_nsp_v1_configurationmanager_proto_depIdxs,
	}.Build()
	File_api_nsp_v1_configurationmanager_proto = out.File
	file_api_nsp_v1_configurationmanager_proto_rawDesc = nil
	file_api_nsp_v1_configurationmanager_proto_goTypes = nil
	file_api_nsp_v1_configurationmanager_proto_depIdxs = nil
}
