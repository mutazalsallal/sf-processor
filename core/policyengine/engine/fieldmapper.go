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
package engine

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/cespare/xxhash"
	"github.com/sysflow-telemetry/sf-apis/go/logger"
	"github.com/sysflow-telemetry/sf-apis/go/sfgo"
)

// FieldMap is a functional type denoting a SysFlow attribute mapper.
type FieldMap func(r *Record) interface{}

// IntFieldMap is a functional type denoting a numerical attribute mapper.
type IntFieldMap func(r *Record) int64

// StrFieldMap is a functional type denoting a string attribute mapper.
type StrFieldMap func(r *Record) string

// FieldMapper is an adapter for SysFlow attribute mappers.
type FieldMapper struct {
	Mappers map[string]FieldMap
}

// Map retrieves a field map based on a SysFlow attribute.
func (m FieldMapper) Map(attr string) FieldMap {
	if mapper, ok := m.Mappers[attr]; ok {
		return mapper
	}
	return func(r *Record) interface{} { return attr }
}

// MapInt retrieves a numerical field map based on a SysFlow attribute.
func (m FieldMapper) MapInt(attr string) IntFieldMap {
	return func(r *Record) int64 {
		if v, ok := m.Map(attr)(r).(int64); ok {
			return v
		} else if v, err := strconv.ParseInt(attr, 10, 64); err == nil {
			return v
		}
		return sfgo.Zeros.Int64
	}
}

// MapStr retrieves a string field map based on a SysFlow attribute.
func (m FieldMapper) MapStr(attr string) StrFieldMap {
	return func(r *Record) string {
		if v, ok := m.Map(attr)(r).(string); ok {
			return trimBoundingQuotes(v)
		} else if v, ok := m.Map(attr)(r).(int64); ok {
			return strconv.FormatInt(v, 10)
		} else if v, ok := m.Map(attr)(r).(bool); ok {
			return strconv.FormatBool(v)
		}
		return sfgo.Zeros.String
	}
}

// Fields defines a sorted array of all exported field mapper keys.
var Fields = getFields()

// Mapper defines a global attribute mapper instance.
var Mapper = FieldMapper{getMappers()}

// getFields returns a sorted array of all exported field mapper keys.
func getFields() []string {
	mappers := getExportedMappers()
	keys := make([]string, 0, len(mappers))
	for k := range mappers {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i int, j int) bool {
		ki := len(strings.Split(keys[i], "."))
		kj := len(strings.Split(keys[j], "."))
		if ki == kj {
			return strings.Compare(keys[i], keys[j]) < 0
		}
		return ki < kj
	})
	return keys
}

func getMappers() map[string]FieldMap {
	mappers := getExportedMappers()
	for k, v := range getNonExportedMappers() {
		if _, ok := mappers[k]; !ok {
			mappers[k] = v
		} else if ok {
			logger.Warn.Println("Duplicate mapper key: ", k)
		}
	}
	return mappers
}

// getExportedMappers defines all mappers for exported attributes.
func getExportedMappers() map[string]FieldMap {
	return map[string]FieldMap{
		// SysFlow
		SF_TYPE:                 mapRecType(sfgo.SYSFLOW_SRC),
		SF_OPFLAGS:              mapOpFlags(sfgo.SYSFLOW_SRC),
		SF_RET:                  mapRet(sfgo.SYSFLOW_SRC),
		SF_TS:                   mapInt(sfgo.SYSFLOW_SRC, sfgo.TS_INT),
		SF_ENDTS:                mapEndTs(sfgo.SYSFLOW_SRC),
		SF_PROC_OID:             mapOID(sfgo.SYSFLOW_SRC, sfgo.PROC_OID_HPID_INT, sfgo.PROC_OID_CREATETS_INT),
		SF_PROC_PID:             mapInt(sfgo.SYSFLOW_SRC, sfgo.PROC_OID_HPID_INT),
		SF_PROC_NAME:            mapName(sfgo.SYSFLOW_SRC, sfgo.PROC_EXE_STR),
		SF_PROC_EXE:             mapStr(sfgo.SYSFLOW_SRC, sfgo.PROC_EXE_STR),
		SF_PROC_ARGS:            mapStr(sfgo.SYSFLOW_SRC, sfgo.PROC_EXEARGS_STR),
		SF_PROC_UID:             mapInt(sfgo.SYSFLOW_SRC, sfgo.PROC_UID_INT),
		SF_PROC_USER:            mapStr(sfgo.SYSFLOW_SRC, sfgo.PROC_USERNAME_STR),
		SF_PROC_TID:             mapInt(sfgo.SYSFLOW_SRC, sfgo.TID_INT),
		SF_PROC_GID:             mapInt(sfgo.SYSFLOW_SRC, sfgo.PROC_GID_INT),
		SF_PROC_GROUP:           mapStr(sfgo.SYSFLOW_SRC, sfgo.PROC_GROUPNAME_STR),
		SF_PROC_CREATETS:        mapInt(sfgo.SYSFLOW_SRC, sfgo.PROC_OID_CREATETS_INT),
		SF_PROC_TTY:             mapInt(sfgo.SYSFLOW_SRC, sfgo.PROC_TTY_INT),
		SF_PROC_ENTRY:           mapEntry(sfgo.SYSFLOW_SRC, sfgo.PROC_ENTRY_INT),
		SF_PROC_CMDLINE:         mapJoin(sfgo.SYSFLOW_SRC, sfgo.PROC_EXE_STR, sfgo.PROC_EXEARGS_STR),
		SF_PROC_ANAME:           mapCachedValue(sfgo.SYSFLOW_SRC, ProcAName),
		SF_PROC_AEXE:            mapCachedValue(sfgo.SYSFLOW_SRC, ProcAExe),
		SF_PROC_ACMDLINE:        mapCachedValue(sfgo.SYSFLOW_SRC, ProcACmdLine),
		SF_PROC_APID:            mapCachedValue(sfgo.SYSFLOW_SRC, ProcAPID),
		SF_PPROC_OID:            mapOID(sfgo.SYSFLOW_SRC, sfgo.PROC_POID_HPID_INT, sfgo.PROC_POID_CREATETS_INT),
		SF_PPROC_PID:            mapInt(sfgo.SYSFLOW_SRC, sfgo.PROC_POID_HPID_INT),
		SF_PPROC_NAME:           mapCachedValue(sfgo.SYSFLOW_SRC, PProcName),
		SF_PPROC_EXE:            mapCachedValue(sfgo.SYSFLOW_SRC, PProcExe),
		SF_PPROC_ARGS:           mapCachedValue(sfgo.SYSFLOW_SRC, PProcArgs),
		SF_PPROC_UID:            mapCachedValue(sfgo.SYSFLOW_SRC, PProcUID),
		SF_PPROC_USER:           mapCachedValue(sfgo.SYSFLOW_SRC, PProcUser),
		SF_PPROC_GID:            mapCachedValue(sfgo.SYSFLOW_SRC, PProcGID),
		SF_PPROC_GROUP:          mapCachedValue(sfgo.SYSFLOW_SRC, PProcGroup),
		SF_PPROC_CREATETS:       mapInt(sfgo.SYSFLOW_SRC, sfgo.PROC_POID_CREATETS_INT),
		SF_PPROC_TTY:            mapCachedValue(sfgo.SYSFLOW_SRC, PProcTTY),
		SF_PPROC_ENTRY:          mapCachedValue(sfgo.SYSFLOW_SRC, PProcEntry),
		SF_PPROC_CMDLINE:        mapCachedValue(sfgo.SYSFLOW_SRC, PProcCmdLine),
		SF_FILE_NAME:            mapName(sfgo.SYSFLOW_SRC, sfgo.FILE_PATH_STR),
		SF_FILE_PATH:            mapPath(sfgo.SYSFLOW_SRC, sfgo.FILE_PATH_STR),
		SF_FILE_SYMLINK:         mapSymlink(sfgo.SYSFLOW_SRC, sfgo.FILE_PATH_STR),
		SF_FILE_OID:             mapStr(sfgo.SYSFLOW_SRC, sfgo.FILE_OID_STR),
		SF_FILE_DIRECTORY:       mapDir(sfgo.SYSFLOW_SRC, sfgo.FILE_PATH_STR),
		SF_FILE_NEWNAME:         mapName(sfgo.SYSFLOW_SRC, sfgo.SEC_FILE_PATH_STR),
		SF_FILE_NEWPATH:         mapPath(sfgo.SYSFLOW_SRC, sfgo.SEC_FILE_PATH_STR),
		SF_FILE_NEWSYMLINK:      mapSymlink(sfgo.SYSFLOW_SRC, sfgo.SEC_FILE_PATH_STR),
		SF_FILE_NEWOID:          mapStr(sfgo.SYSFLOW_SRC, sfgo.SEC_FILE_OID_STR),
		SF_FILE_NEWDIRECTORY:    mapDir(sfgo.SYSFLOW_SRC, sfgo.SEC_FILE_PATH_STR),
		SF_FILE_TYPE:            mapFileType(sfgo.SYSFLOW_SRC, sfgo.FILE_RESTYPE_INT),
		SF_FILE_IS_OPEN_WRITE:   mapIsOpenWrite(sfgo.SYSFLOW_SRC, sfgo.FL_FILE_OPENFLAGS_INT),
		SF_FILE_IS_OPEN_READ:    mapIsOpenRead(sfgo.SYSFLOW_SRC, sfgo.FL_FILE_OPENFLAGS_INT),
		SF_FILE_FD:              mapInt(sfgo.SYSFLOW_SRC, sfgo.FL_FILE_FD_INT),
		SF_FILE_OPENFLAGS:       mapOpenFlags(sfgo.SYSFLOW_SRC, sfgo.FL_FILE_OPENFLAGS_INT),
		SF_NET_PROTO:            mapInt(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_PROTO_INT),
		SF_NET_SPORT:            mapInt(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_SPORT_INT),
		SF_NET_DPORT:            mapInt(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_DPORT_INT),
		SF_NET_PORT:             mapPort(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_SPORT_INT, sfgo.FL_NETW_DPORT_INT),
		SF_NET_SIP:              mapIP(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_SIP_INT),
		SF_NET_DIP:              mapIP(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_DIP_INT),
		SF_NET_IP:               mapIP(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_SIP_INT, sfgo.FL_NETW_DIP_INT),
		SF_FLOW_RBYTES:          mapSum(sfgo.SYSFLOW_SRC, sfgo.FL_FILE_NUMRRECVBYTES_INT, sfgo.FL_NETW_NUMRRECVBYTES_INT),
		SF_FLOW_ROPS:            mapSum(sfgo.SYSFLOW_SRC, sfgo.FL_FILE_NUMRRECVOPS_INT, sfgo.FL_NETW_NUMRRECVOPS_INT),
		SF_FLOW_WBYTES:          mapSum(sfgo.SYSFLOW_SRC, sfgo.FL_FILE_NUMWSENDBYTES_INT, sfgo.FL_NETW_NUMWSENDBYTES_INT),
		SF_FLOW_WOPS:            mapSum(sfgo.SYSFLOW_SRC, sfgo.FL_FILE_NUMWSENDOPS_INT, sfgo.FL_NETW_NUMWSENDOPS_INT),
		SF_CONTAINER_ID:         mapStr(sfgo.SYSFLOW_SRC, sfgo.CONT_ID_STR),
		SF_CONTAINER_NAME:       mapStr(sfgo.SYSFLOW_SRC, sfgo.CONT_NAME_STR),
		SF_CONTAINER_IMAGEID:    mapStr(sfgo.SYSFLOW_SRC, sfgo.CONT_IMAGEID_STR),
		SF_CONTAINER_IMAGE:      mapStr(sfgo.SYSFLOW_SRC, sfgo.CONT_IMAGE_STR),
		SF_CONTAINER_TYPE:       mapContType(sfgo.SYSFLOW_SRC, sfgo.CONT_TYPE_INT),
		SF_CONTAINER_PRIVILEGED: mapInt(sfgo.SYSFLOW_SRC, sfgo.CONT_PRIVILEGED_INT),
		SF_NODE_ID:              mapStr(sfgo.SYSFLOW_SRC, sfgo.SFHE_EXPORTER_STR),
		SF_NODE_IP:              mapStr(sfgo.SYSFLOW_SRC, sfgo.SFHE_IP_STR),
		SF_SCHEMA_VERSION:       mapInt(sfgo.SYSFLOW_SRC, sfgo.SFHE_VERSION_INT),
	}
}

// getExtendedMappers defines all mappers for extended attributes.
func getExtendedMappers() map[string]FieldMap {
	return map[string]FieldMap{
		//Ext processes
		EXT_PROC_GUID_STR:                mapStr(sfgo.PROCESS_SRC, sfgo.PROC_GUID_STR),
		EXT_PROC_IMAGE_STR:               mapStr(sfgo.PROCESS_SRC, sfgo.PROC_IMAGE_STR),
		EXT_PROC_CURR_DIRECTORY_STR:      mapDir(sfgo.PROCESS_SRC, sfgo.PROC_CURR_DIRECTORY_STR),
		EXT_PROC_LOGON_GUID_STR:          mapStr(sfgo.PROCESS_SRC, sfgo.PROC_LOGON_GUID_STR),
		EXT_PROC_LOGON_ID_STR:            mapStr(sfgo.PROCESS_SRC, sfgo.PROC_LOGON_ID_STR),
		EXT_PROC_TERMINAL_SESSION_ID_STR: mapStr(sfgo.PROCESS_SRC, sfgo.PROC_TERMINAL_SESSION_ID_STR),
		EXT_PROC_INTEGRITY_LEVEL_STR:     mapStr(sfgo.PROCESS_SRC, sfgo.PROC_INTEGRITY_LEVEL_STR),
		EXT_PROC_SIGNATURE_STR:           mapStr(sfgo.PROCESS_SRC, sfgo.PROC_SIGNATURE_STR),
		EXT_PROC_SIGNATURE_STATUS_STR:    mapStr(sfgo.PROCESS_SRC, sfgo.PROC_SIGNATURE_STATUS_STR),
		EXT_PROC_SHA1_HASH_STR:           mapStr(sfgo.PROCESS_SRC, sfgo.PROC_SHA1_HASH_STR),
		EXT_PROC_MD5_HASH_STR:            mapStr(sfgo.PROCESS_SRC, sfgo.PROC_MD5_HASH_STR),
		EXT_PROC_SHA256_HASH_STR:         mapStr(sfgo.PROCESS_SRC, sfgo.PROC_SHA256_HASH_STR),
		EXT_PROC_IMP_HASH_STR:            mapStr(sfgo.PROCESS_SRC, sfgo.PROC_IMP_HASH_STR),
		EXT_PROC_SIGNED_INT:              mapInt(sfgo.PROCESS_SRC, sfgo.PROC_SIGNED_INT),

		//Ext files
		EXT_FILE_SIGNATURE_STR:        mapStr(sfgo.FILE_SRC, sfgo.FILE_SIGNATURE_STR),
		EXT_FILE_SIGNATURE_STATUS_STR: mapStr(sfgo.FILE_SRC, sfgo.FILE_SIGNATURE_STATUS_STR),
		EXT_FILE_SHA1_HASH_STR:        mapStr(sfgo.FILE_SRC, sfgo.FILE_SHA1_HASH_STR),
		EXT_FILE_MD5_HASH_STR:         mapStr(sfgo.FILE_SRC, sfgo.FILE_MD5_HASH_STR),
		EXT_FILE_SHA256_HASH_STR:      mapStr(sfgo.FILE_SRC, sfgo.FILE_SHA256_HASH_STR),
		EXT_FILE_IMP_HASH_STR:         mapStr(sfgo.FILE_SRC, sfgo.FILE_IMP_HASH_STR),
		EXT_FILE_SIGNED_INT:           mapInt(sfgo.FILE_SRC, sfgo.FILE_SIGNED_INT),

		//Ext network
		EXT_NET_SOURCE_HOST_NAME_STR: mapStr(sfgo.NETWORK_SRC, sfgo.NET_SOURCE_HOST_NAME_STR),
		EXT_NET_SOURCE_PORT_NAME_STR: mapStr(sfgo.NETWORK_SRC, sfgo.NET_SOURCE_PORT_NAME_STR),
		EXT_NET_DEST_HOST_NAME_STR:   mapStr(sfgo.NETWORK_SRC, sfgo.NET_DEST_HOST_NAME_STR),
		EXT_NET_DEST_PORT_NAME_STR:   mapStr(sfgo.NETWORK_SRC, sfgo.NET_DEST_PORT_NAME_STR),

		//Ext target proc
		EXT_TARG_PROC_OID_CREATETS_INT:       mapInt(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_OID_CREATETS_INT),
		EXT_TARG_PROC_OID_HPID_INT:           mapInt(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_OID_HPID_INT),
		EXT_TARG_PROC_TS_INT:                 mapInt(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_TS_INT),
		EXT_TARG_PROC_POID_CREATETS_INT:      mapInt(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_POID_CREATETS_INT),
		EXT_TARG_PROC_POID_HPID_INT:          mapInt(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_POID_HPID_INT),
		EXT_TARG_PROC_EXE_STR:                mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_EXE_STR),
		EXT_TARG_PROC_EXEARGS_STR:            mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_EXEARGS_STR),
		EXT_TARG_PROC_UID_INT:                mapInt(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_UID_INT),
		EXT_TARG_PROC_GID_INT:                mapInt(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_GID_INT),
		EXT_TARG_PROC_USERNAME_STR:           mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_USERNAME_STR),
		EXT_TARG_PROC_GROUPNAME_STR:          mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_GROUPNAME_STR),
		EXT_TARG_PROC_TTY_INT:                mapInt(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_TTY_INT),
		EXT_TARG_PROC_CONTAINERID_STRING_STR: mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_CONTAINERID_STRING_STR),
		EXT_TARG_PROC_ENTRY_INT:              mapEntry(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_ENTRY_INT),

		EXT_TARG_PROC_GUID_STR:                mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_GUID_STR),
		EXT_TARG_PROC_IMAGE_STR:               mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_IMAGE_STR),
		EXT_TARG_PROC_CURR_DIRECTORY_STR:      mapDir(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_CURR_DIRECTORY_STR),
		EXT_TARG_PROC_LOGON_GUID_STR:          mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_LOGON_GUID_STR),
		EXT_TARG_PROC_LOGON_ID_STR:            mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_LOGON_ID_STR),
		EXT_TARG_PROC_TERMINAL_SESSION_ID_STR: mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_TERMINAL_SESSION_ID_STR),
		EXT_TARG_PROC_INTEGRITY_LEVEL_STR:     mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_INTEGRITY_LEVEL_STR),
		EXT_TARG_PROC_SIGNATURE_STR:           mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_SIGNATURE_STR),
		EXT_TARG_PROC_SIGNATURE_STATUS_STR:    mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_SIGNATURE_STATUS_STR),
		EXT_TARG_PROC_SHA1_HASH_STR:           mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_SHA1_HASH_STR),
		EXT_TARG_PROC_MD5_HASH_STR:            mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_MD5_HASH_STR),
		EXT_TARG_PROC_SHA256_HASH_STR:         mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_SHA256_HASH_STR),
		EXT_TARG_PROC_IMP_HASH_STR:            mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_IMP_HASH_STR),
		EXT_TARG_PROC_SIGNED_INT:              mapInt(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_SIGNED_INT),
		EXT_TARG_PROC_START_ADDR_STR:          mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_START_ADDR_STR),
		EXT_TARG_PROC_START_MODULE_STR:        mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_START_MODULE_STR),
		EXT_TARG_PROC_START_FUNCTION_STR:      mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_START_FUNCTION_STR),
		EXT_TARG_PROC_GRANT_ACCESS_STR:        mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_GRANT_ACCESS_STR),
		EXT_TARG_PROC_CALL_TRACE_STR:          mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_CALL_TRACE_STR),
		EXT_TARG_PROC_ACCESS_TYPE_STR:         mapStr(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_ACCESS_TYPE_STR),
		EXT_TARG_PROC_NEW_THREAD_ID_INT:       mapInt(sfgo.TARG_PROC_SRC, sfgo.EVT_TARG_PROC_NEW_THREAD_ID_INT),
	}
}

// getNonExportedMappers defines all mappers for non-exported (query-only) attributes.
func getNonExportedMappers() map[string]FieldMap {
	return map[string]FieldMap{
		// Falco
		FALCO_EVT_TYPE:              mapEvtType(sfgo.SYSFLOW_SRC),
		FALCO_EVT_RAW_RES:           mapRecType(sfgo.SYSFLOW_SRC),
		FALCO_EVT_RAW_TIME:          mapInt(sfgo.SYSFLOW_SRC, sfgo.TS_INT),
		FALCO_EVT_DIR:               mapConsts(FALCO_ENTER_EVENT, FALCO_EXIT_EVENT),
		FALCO_EVT_IS_OPEN_READ:      mapIsOpenRead(sfgo.SYSFLOW_SRC, sfgo.FL_FILE_OPENFLAGS_INT),
		FALCO_EVT_IS_OPEN_WRITE:     mapIsOpenWrite(sfgo.SYSFLOW_SRC, sfgo.FL_FILE_OPENFLAGS_INT),
		FALCO_EVT_UID:               mapInt(sfgo.SYSFLOW_SRC, sfgo.PROC_UID_INT),
		FALCO_EVT_NAME:              mapName(sfgo.SYSFLOW_SRC, sfgo.FILE_PATH_STR),
		FALCO_EVT_PATH:              mapPath(sfgo.SYSFLOW_SRC, sfgo.FILE_PATH_STR),
		FALCO_EVT_NEWPATH:           mapPath(sfgo.SYSFLOW_SRC, sfgo.SEC_FILE_PATH_STR),
		FALCO_EVT_OLDPATH:           mapPath(sfgo.SYSFLOW_SRC, sfgo.FILE_PATH_STR),
		FALCO_FD_TYPECHAR:           mapFileType(sfgo.SYSFLOW_SRC, sfgo.FILE_RESTYPE_INT),
		FALCO_FD_DIRECTORY:          mapDir(sfgo.SYSFLOW_SRC, sfgo.FILE_PATH_STR),
		FALCO_FD_NAME:               mapName(sfgo.SYSFLOW_SRC, sfgo.FILE_PATH_STR),
		FALCO_FD_FILENAME:           mapName(sfgo.SYSFLOW_SRC, sfgo.FILE_PATH_STR),
		FALCO_FD_PROTO:              mapProto(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_PROTO_INT),
		FALCO_FD_LPROTO:             mapProto(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_PROTO_INT),
		FALCO_FD_L4PROTO:            mapProto(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_PROTO_INT),
		FALCO_FD_RPROTO:             mapProto(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_PROTO_INT),
		FALCO_FD_SPROTO:             mapProto(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_PROTO_INT),
		FALCO_FD_CPROTO:             mapProto(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_PROTO_INT),
		FALCO_FD_SPORT:              mapInt(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_SPORT_INT),
		FALCO_FD_DPORT:              mapInt(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_DPORT_INT),
		FALCO_FD_SIP:                mapIP(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_SIP_INT),
		FALCO_FD_DIP:                mapIP(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_DIP_INT),
		FALCO_FD_IP:                 mapIP(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_SIP_INT, sfgo.FL_NETW_DIP_INT),
		FALCO_FD_PORT:               mapPort(sfgo.SYSFLOW_SRC, sfgo.FL_NETW_SPORT_INT, sfgo.FL_NETW_DPORT_INT),
		FALCO_FD_NUM:                mapInt(sfgo.SYSFLOW_SRC, sfgo.FL_FILE_FD_INT),
		FALCO_USER_NAME:             mapStr(sfgo.SYSFLOW_SRC, sfgo.PROC_USERNAME_STR),
		FALCO_PROC_PID:              mapInt(sfgo.SYSFLOW_SRC, sfgo.PROC_OID_HPID_INT),
		FALCO_PROC_TID:              mapInt(sfgo.SYSFLOW_SRC, sfgo.TID_INT),
		FALCO_PROC_GID:              mapInt(sfgo.SYSFLOW_SRC, sfgo.PROC_GID_INT),
		FALCO_PROC_UID:              mapInt(sfgo.SYSFLOW_SRC, sfgo.PROC_UID_INT),
		FALCO_PROC_GROUP:            mapStr(sfgo.SYSFLOW_SRC, sfgo.PROC_GROUPNAME_STR),
		FALCO_PROC_TTY:              mapCachedValue(sfgo.SYSFLOW_SRC, PProcTTY),
		FALCO_PROC_USER:             mapStr(sfgo.SYSFLOW_SRC, sfgo.PROC_USERNAME_STR),
		FALCO_PROC_EXE:              mapStr(sfgo.SYSFLOW_SRC, sfgo.PROC_EXE_STR),
		FALCO_PROC_NAME:             mapName(sfgo.SYSFLOW_SRC, sfgo.PROC_EXE_STR),
		FALCO_PROC_ARGS:             mapStr(sfgo.SYSFLOW_SRC, sfgo.PROC_EXEARGS_STR),
		FALCO_PROC_CREATE_TIME:      mapInt(sfgo.SYSFLOW_SRC, sfgo.PROC_POID_CREATETS_INT),
		FALCO_PROC_CMDLINE:          mapJoin(sfgo.SYSFLOW_SRC, sfgo.PROC_EXE_STR, sfgo.PROC_EXEARGS_STR),
		FALCO_PROC_ANAME:            mapCachedValue(sfgo.SYSFLOW_SRC, ProcAName),
		FALCO_PROC_APID:             mapCachedValue(sfgo.SYSFLOW_SRC, ProcAPID),
		FALCO_PROC_PPID:             mapInt(sfgo.SYSFLOW_SRC, sfgo.PROC_POID_HPID_INT),
		FALCO_PROC_PGID:             mapCachedValue(sfgo.SYSFLOW_SRC, PProcGID),
		FALCO_PROC_PUID:             mapCachedValue(sfgo.SYSFLOW_SRC, PProcUID),
		FALCO_PROC_PGROUP:           mapCachedValue(sfgo.SYSFLOW_SRC, PProcGroup),
		FALCO_PROC_PTTY:             mapCachedValue(sfgo.SYSFLOW_SRC, PProcTTY),
		FALCO_PROC_PUSER:            mapCachedValue(sfgo.SYSFLOW_SRC, PProcUser),
		FALCO_PROC_PEXE:             mapCachedValue(sfgo.SYSFLOW_SRC, PProcExe),
		FALCO_PROC_PARGS:            mapCachedValue(sfgo.SYSFLOW_SRC, PProcArgs),
		FALCO_PROC_PCREATE_TIME:     mapInt(sfgo.SYSFLOW_SRC, sfgo.PROC_POID_CREATETS_INT),
		FALCO_PROC_PNAME:            mapCachedValue(sfgo.SYSFLOW_SRC, PProcName),
		FALCO_PROC_PCMDLINE:         mapCachedValue(sfgo.SYSFLOW_SRC, PProcCmdLine),
		FALCO_CONT_ID:               mapStr(sfgo.SYSFLOW_SRC, sfgo.CONT_ID_STR),
		FALCO_CONT_IMAGE_ID:         mapStr(sfgo.SYSFLOW_SRC, sfgo.CONT_IMAGEID_STR),
		FALCO_CONT_IMAGE_REPOSITORY: mapRepo(sfgo.SYSFLOW_SRC, sfgo.CONT_IMAGE_STR),
		FALCO_CONT_IMAGE:            mapStr(sfgo.SYSFLOW_SRC, sfgo.CONT_IMAGE_STR),
		FALCO_CONT_NAME:             mapStr(sfgo.SYSFLOW_SRC, sfgo.CONT_NAME_STR),
		FALCO_CONT_TYPE:             mapContType(sfgo.SYSFLOW_SRC, sfgo.CONT_TYPE_INT),
		FALCO_CONT_PRIVILEGED:       mapInt(sfgo.SYSFLOW_SRC, sfgo.CONT_PRIVILEGED_INT),
	}
}

func mapStr(src sfgo.Source, attr sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} { return r.GetStr(attr, src) }
}

func mapInt(src sfgo.Source, attr sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} { return r.GetInt(attr, src) }
}

func mapSum(src sfgo.Source, attrs ...sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		var sum int64 = 0
		for _, attr := range attrs {
			sum += r.GetInt(attr, src)
		}
		return sum
	}
}

func mapJoin(src sfgo.Source, attrs ...sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		var join string = r.GetStr(attrs[0], src)
		for _, attr := range attrs[1:] {
			join += SPACE + r.GetStr(attr, src)
		}
		return join
	}
}

func mapRecType(src sfgo.Source) FieldMap {
	return func(r *Record) interface{} {
		switch r.GetInt(sfgo.SF_REC_TYPE, src) {
		case sfgo.PROC:
			return TyP
		case sfgo.FILE:
			return TyF
		case sfgo.CONT:
			return TyC
		case sfgo.PROC_EVT:
			return TyPE
		case sfgo.FILE_EVT:
			return TyFE
		case sfgo.FILE_FLOW:
			return TyFF
		case sfgo.NET_FLOW:
			return TyNF
		case sfgo.HEADER:
			return TyH
		default:
			return TyUnknow
		}
	}
}

func mapOpFlags(src sfgo.Source) FieldMap {
	return func(r *Record) interface{} {
		opflags := r.GetInt(sfgo.EV_PROC_OPFLAGS_INT, src)
		rtype := mapRecType(src)(r).(string)
		return strings.Join(sfgo.GetOpFlags(int32(opflags), rtype), LISTSEP)
	}
}

func mapEvtType(src sfgo.Source) FieldMap {
	return func(r *Record) interface{} {
		opflags := r.GetInt(sfgo.EV_PROC_OPFLAGS_INT, src)
		rtype := mapRecType(src)(r).(string)
		return strings.Join(sfgo.GetEvtTypes(int32(opflags), rtype), LISTSEP)
	}
}

func mapRet(src sfgo.Source) FieldMap {
	return func(r *Record) interface{} {
		switch r.GetInt(sfgo.SF_REC_TYPE, src) {
		case sfgo.PROC_EVT:
			fallthrough
		case sfgo.FILE_EVT:
			return r.GetInt(sfgo.RET_INT, src)
		default:
			return sfgo.Zeros.Int64
		}
	}
}

func mapEndTs(src sfgo.Source) FieldMap {
	return func(r *Record) interface{} {
		switch r.GetInt(sfgo.SF_REC_TYPE, src) {
		case sfgo.FILE_FLOW:
			return r.GetInt(sfgo.FL_FILE_ENDTS_INT, src)
		case sfgo.NET_FLOW:
			return r.GetInt(sfgo.FL_NETW_ENDTS_INT, src)
		default:
			return sfgo.Zeros.Int64
		}
	}
}

func mapEntry(src sfgo.Source, attr sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		if r.GetInt(attr, src) == 1 {
			return true
		}
		return false
	}
}

func mapName(src sfgo.Source, attr sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		return filepath.Base(r.GetStr(attr, src))
	}
}

func mapDir(src sfgo.Source, attr sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		return filepath.Dir(mapPath(src, attr)(r).(string))
	}
}

func mapRepo(src sfgo.Source, attr sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		return strings.Split(r.GetStr(attr, src), ":")[0]
	}
}

func mapPath(src sfgo.Source, attr sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		path, _ := parseSymPath(src, attr, r)
		return path
	}
}

func mapSymlink(src sfgo.Source, attr sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		_, link := parseSymPath(src, attr, r)
		return link
	}
}

func mapFileType(src sfgo.Source, attr sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		return sfgo.GetFileType(r.GetInt(attr, src))
	}
}

func mapIsOpenWrite(src sfgo.Source, attr sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		if sfgo.IsOpenWrite(r.GetInt(attr, src)) {
			return true
		}
		return false
	}
}

func mapIsOpenRead(src sfgo.Source, attr sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		if sfgo.IsOpenRead(r.GetInt(attr, src)) {
			return true
		}
		return false
	}
}

func mapOpenFlags(src sfgo.Source, attr sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		return strings.Join(sfgo.GetOpenFlags(r.GetInt(attr, src)), LISTSEP)
	}
}

func mapProto(src sfgo.Source, attr sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		return sfgo.GetProto(r.GetInt(attr, src))
	}
}

func mapPort(src sfgo.Source, attrs ...sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		var ports = make([]string, 0)
		for _, attr := range attrs {
			ports = append(ports, strconv.FormatInt(r.GetInt(attr, src), 10))
		}
		// logger.Info.Println(ports)
		return strings.Join(ports, LISTSEP)
	}
}

func mapIP(src sfgo.Source, attrs ...sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		var ips = make([]string, 0)
		for _, attr := range attrs {
			ips = append(ips, sfgo.GetIPStr(int32(r.GetInt(attr, src))))
		}
		// logger.Info.Println(ips)
		return strings.Join(ips, LISTSEP)
	}
}

func mapContType(src sfgo.Source, attr sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		return sfgo.GetContType(r.GetInt(attr, src))
	}
}

func mapCachedValue(src sfgo.Source, attr RecAttribute) FieldMap {
	return func(r *Record) interface{} {
		oid := sfgo.OID{CreateTS: r.GetInt(sfgo.PROC_OID_CREATETS_INT, src), Hpid: r.GetInt(sfgo.PROC_OID_HPID_INT, src)}
		return r.GetCachedValue(oid, attr)
	}
}

func mapOID(src sfgo.Source, attrs ...sfgo.Attribute) FieldMap {
	return func(r *Record) interface{} {
		h := xxhash.New()
		for _, attr := range attrs {
			h.Write([]byte(fmt.Sprintf("%v", r.GetInt(attr, src))))
		}
		return fmt.Sprintf("%x", h.Sum(nil))
	}
}

func mapConsts(consts ...string) FieldMap {
	return func(r *Record) interface{} {
		return strings.Join(consts, LISTSEP)
	}
}

func mapNa(attr string) FieldMap {
	return func(r *Record) interface{} {
		logger.Warn.Println("Attribute not supported ", attr)
		return sfgo.Zeros.String
	}
}
