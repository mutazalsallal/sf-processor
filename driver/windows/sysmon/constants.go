//
// Copyright (C) 2020 IBM Corporation.
//
// Authors:
// Frederico Araujo <frederico.araujo@ibm.com>
// Teryl Taylor <terylt@ibm.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package sysmon

const (
	cSysmonProcessCreate              = 1
	cSysmonNetworkConnection          = 3
	cSysmonProcessExit                = 5
	cSysmonLoadImage                  = 7
	cSysmonCreateRemoteThread         = 8
	cSysmonProcessAccess              = 10
	cSysmonFileCreated                = 11
	cSysmonCreateDeleteRegistryObject = 12
	cSysmonSetRegistryValue           = 13
	cSysmonPipeCreated                = 17
	cSysmonPipeConnected              = 18
	cEvtLogProvider                   = "Microsoft-Windows-Sysmon/Operational"

	cUtcTime           = "UtcTime"
	cProcessGUID       = "ProcessGuid"
	cProcessID         = "ProcessId"
	cUser              = "User"
	cImage             = "Image"
	cCommandLine       = "CommandLine"
	cCurrentDirectory  = "CurrentDirectory"
	cLogonGUID         = "LogonGuid"
	cLogonID           = "LogonId"
	cTerminalSessionID = "TerminalSessionId"
	cIntegrityLevel    = "IntegrityLevel"
	cHashes            = "Hashes"
	cParentProcessGUID = "ParentProcessGuid"
	cParentProcessID   = "ParentProcessId"
	cParentImage       = "ParentImage"
	cParentCommandLine = "ParentCommandLine"

	cImageLoaded     = "ImageLoaded"
	cSigned          = "Signed"
	cSignature       = "Signature"
	cSignatureStatus = "SignatureStatus"

	cProtocol            = "Protocol"
	cInitiated           = "Initiated"
	cSourceIsIpv6        = "SourceIsIpv6"
	cSourceIP            = "SourceIp"
	cSourceHostname      = "SourceHostname"
	cSourcePort          = "SourcePort"
	cSourcePortName      = "SourcePortName"
	cDestinationIsIpv6   = "DestinationIsIpv6"
	cDestinationIP       = "DestinationIp"
	cDestinationHostname = "DestinationHostname"
	cDestinationPort     = "DestinationPort"
	cDestinationPortName = "DestinationPortName"

	cTargetFilename  = "TargetFilename"
	cCreationUtcTime = "CreationUtcTime"

	cTargetObject = "TargetObject"
	cDetails      = "Details"

	cSourceProcessGUID = "SourceProcessGuid"
	cSourceProcessID   = "SourceProcessId"
	cSourceImage       = "SourceImage"
	cSourceThreadID    = "SourceThreadId"
	cTargetProcessGUID = "TargetProcessGuid"
	cTargetProcessID   = "TargetProcessId"
	cTargetImage       = "TargetImage"
	cNewThreadID       = "NewThreadId"
	cStartAddress      = "StartAddress"
	cStartModule       = "StartModule"
	cStartFunction     = "StartFunction"
	cGrantedAccess     = "GrantedAccess"
	cCallTrace         = "CallTrace"

	cEventType = "EventType"

	cTimeFormat = "2006-01-02 15:04:05.000"

	cHashRegex = "^SHA1=([A-Z0-9]+),MD5=([A-Z0-9]+),SHA256=([A-Z0-9]+),IMPHASH=([A-Z0-9]+)$"

	cDeleteValue = "DeleteValue"
	cSetValue    = "SetValue"
)