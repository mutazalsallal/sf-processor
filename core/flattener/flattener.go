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
package flattener

import (
	"encoding/hex"

	"github.com/sysflow-telemetry/sf-apis/go/logger"
	"github.com/sysflow-telemetry/sf-apis/go/plugins"
	"github.com/sysflow-telemetry/sf-apis/go/sfgo"
)

const (
	handlerName string = "flattener"
	channelName string = "flattenerchan"
)

// FlatChannel defines a multi-source flat channel
type FlatChannel struct {
	In chan *sfgo.FlatRecord
}

// NewFlattenerChan creates a new channel with given capacity.
func NewFlattenerChan(size int) interface{} {
	return &FlatChannel{In: make(chan *sfgo.FlatRecord, size)}
}

// Flattener defines the main class for the flatterner plugin.
type Flattener struct {
	outCh chan *sfgo.FlatRecord
}

// NewFlattener creates a new Flattener instance.
func NewFlattener() plugins.SFHandler {
	return new(Flattener)
}

// Register registers plugin to plugin cache.
func (s *Flattener) Register(pc plugins.SFPluginCache) {
	pc.AddHandler(handlerName, NewFlattener)
	pc.AddChannel(channelName, NewFlattenerChan)
}

// Init initializes the handler with a configuration map.
func (s *Flattener) Init(conf map[string]string) error {
	return nil
}

// IsEntityEnabled is used to check if the flattener returns entity records.
func (s *Flattener) IsEntityEnabled() bool {
	return false
}

// SetOutChan sets the plugin output channel.
func (s *Flattener) SetOutChan(chObj interface{}) {
	s.outCh = chObj.(*FlatChannel).In
}

// Cleanup tears down resources.
func (s *Flattener) Cleanup() {
	logger.Trace.Println("Calling Cleanup on Flattener channel")
	if s.outCh != nil {
		close(s.outCh)
	}
}

// HandleHeader processes Header entities.
func (s *Flattener) HandleHeader(hdr *sfgo.SFHeader) error {
	return nil
}

// HandleContainer processes Container entities.
func (s *Flattener) HandleContainer(hdr *sfgo.SFHeader, cont *sfgo.Container) error {
	return nil
}

// HandleProcess processes Process entities.
func (s *Flattener) HandleProcess(hdr *sfgo.SFHeader, cont *sfgo.Container, proc *sfgo.Process) error {
	return nil
}

// HandleFile processes File entities.
func (s *Flattener) HandleFile(hdr *sfgo.SFHeader, cont *sfgo.Container, file *sfgo.File) error {
	return nil
}

// HandleNetFlow processes Network Flows.
func (s *Flattener) HandleNetFlow(hdr *sfgo.SFHeader, cont *sfgo.Container, proc *sfgo.Process, nf *sfgo.NetworkFlow) error {
	fr := newFlatRecord()
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.SF_REC_TYPE] = sfgo.NET_FLOW
	s.fillEntities(hdr, cont, proc, nil, fr)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_NETW_TS_INT] = nf.Ts
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_NETW_TID_INT] = nf.Tid
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_NETW_OPFLAGS_INT] = int64(nf.OpFlags)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_NETW_ENDTS_INT] = nf.EndTs
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_NETW_SIP_INT] = int64(nf.Sip)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_NETW_SPORT_INT] = int64(nf.Sport)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_NETW_DIP_INT] = int64(nf.Dip)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_NETW_DPORT_INT] = int64(nf.Dport)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_NETW_PROTO_INT] = int64(nf.Proto)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_NETW_FD_INT] = int64(nf.Fd)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_NETW_NUMRRECVOPS_INT] = nf.NumRRecvOps
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_NETW_NUMWSENDOPS_INT] = nf.NumWSendOps
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_NETW_NUMRRECVBYTES_INT] = nf.NumRRecvBytes
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_NETW_NUMWSENDBYTES_INT] = nf.NumWSendBytes
	s.outCh <- fr
	return nil
}

// HandleFileFlow processes File Flows.
func (s *Flattener) HandleFileFlow(hdr *sfgo.SFHeader, cont *sfgo.Container, proc *sfgo.Process, file *sfgo.File, ff *sfgo.FileFlow) error {
	fr := newFlatRecord()
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.SF_REC_TYPE] = sfgo.FILE_FLOW
	s.fillEntities(hdr, cont, proc, file, fr)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_FILE_TS_INT] = ff.Ts
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_FILE_TID_INT] = ff.Tid
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_FILE_OPFLAGS_INT] = int64(ff.OpFlags)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_FILE_OPENFLAGS_INT] = int64(ff.OpenFlags)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_FILE_ENDTS_INT] = ff.EndTs
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_FILE_FD_INT] = int64(ff.Fd)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_FILE_NUMRRECVOPS_INT] = ff.NumRRecvOps
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_FILE_NUMWSENDOPS_INT] = ff.NumWSendOps
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_FILE_NUMRRECVBYTES_INT] = ff.NumRRecvBytes
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FL_FILE_NUMWSENDBYTES_INT] = ff.NumWSendBytes
	s.outCh <- fr
	return nil
}

// HandleFileEvt processes File Events.
func (s *Flattener) HandleFileEvt(hdr *sfgo.SFHeader, cont *sfgo.Container, proc *sfgo.Process, file1 *sfgo.File, file2 *sfgo.File, fe *sfgo.FileEvent) error {
	fr := newFlatRecord()
	if file2 != nil {
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.SEC_FILE_STATE_INT] = int64(file2.State)
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.SEC_FILE_TS_INT] = file2.Ts
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.SEC_FILE_RESTYPE_INT] = int64(file2.Restype)
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.SEC_FILE_PATH_STR] = file2.Path
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.SEC_FILE_OID_STR] = getOIDStr(file2.Oid[:])
		if file2.ContainerId != nil && file2.ContainerId.UnionType == sfgo.UnionNullStringTypeEnumString {
			fr.Strs[sfgo.SYSFLOW_IDX][sfgo.SEC_FILE_CONTAINERID_STRING_STR] = file2.ContainerId.String
		} else {
			fr.Strs[sfgo.SYSFLOW_IDX][sfgo.SEC_FILE_CONTAINERID_STRING_STR] = sfgo.Zeros.String
		}
	} else {
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.SEC_FILE_STATE_INT] = sfgo.Zeros.Int64
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.SEC_FILE_TS_INT] = sfgo.Zeros.Int64
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.SEC_FILE_RESTYPE_INT] = sfgo.Zeros.Int64
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.SEC_FILE_PATH_STR] = sfgo.Zeros.String
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.SEC_FILE_CONTAINERID_STRING_STR] = sfgo.Zeros.String
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.SEC_FILE_OID_STR] = sfgo.Zeros.String
	}
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.SF_REC_TYPE] = sfgo.FILE_EVT
	s.fillEntities(hdr, cont, proc, file1, fr)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.EV_FILE_TS_INT] = fe.Ts
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.EV_FILE_TID_INT] = fe.Tid
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.EV_FILE_OPFLAGS_INT] = int64(fe.OpFlags)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.EV_FILE_RET_INT] = int64(fe.Ret)
	s.outCh <- fr
	return nil
}

// HandleProcEvt processes Process Events.
func (s *Flattener) HandleProcEvt(hdr *sfgo.SFHeader, cont *sfgo.Container, proc *sfgo.Process, pe *sfgo.ProcessEvent) error {
	fr := newFlatRecord()
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.SF_REC_TYPE] = sfgo.PROC_EVT
	s.fillEntities(hdr, cont, proc, nil, fr)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.EV_PROC_TS_INT] = pe.Ts
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.EV_PROC_TID_INT] = pe.Tid
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.EV_PROC_OPFLAGS_INT] = int64(pe.OpFlags)
	fr.Ints[sfgo.SYSFLOW_IDX][sfgo.EV_PROC_RET_INT] = int64(pe.Ret)
	s.outCh <- fr
	return nil
}

func (s *Flattener) fillEntities(hdr *sfgo.SFHeader, cont *sfgo.Container, proc *sfgo.Process, file *sfgo.File, fr *sfgo.FlatRecord) {
	if hdr != nil {
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.SFHE_VERSION_INT] = hdr.Version
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.SFHE_EXPORTER_STR] = hdr.Exporter
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.SFHE_IP_STR] = hdr.Ip
	} else {
		logger.Warn.Println("Event does not have a related header.  This should not happen.")
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.SFHE_VERSION_INT] = sfgo.Zeros.Int64
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.SFHE_EXPORTER_STR] = sfgo.Zeros.String
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.SFHE_IP_STR] = sfgo.Zeros.String
	}
	if cont != nil {
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.CONT_ID_STR] = cont.Id
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.CONT_NAME_STR] = cont.Name
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.CONT_IMAGE_STR] = cont.Image
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.CONT_IMAGEID_STR] = cont.Imageid
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.CONT_TYPE_INT] = int64(cont.Type)
		if cont.Privileged {
			fr.Ints[sfgo.SYSFLOW_IDX][sfgo.CONT_PRIVILEGED_INT] = 1
		} else {
			fr.Ints[sfgo.SYSFLOW_IDX][sfgo.CONT_PRIVILEGED_INT] = 0
		}
	} else {
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.CONT_ID_STR] = sfgo.Zeros.String
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.CONT_NAME_STR] = sfgo.Zeros.String
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.CONT_IMAGE_STR] = sfgo.Zeros.String
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.CONT_IMAGEID_STR] = sfgo.Zeros.String
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.CONT_TYPE_INT] = sfgo.Zeros.Int64
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.CONT_PRIVILEGED_INT] = sfgo.Zeros.Int64

	}
	if proc != nil {
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_STATE_INT] = int64(proc.State)
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_OID_CREATETS_INT] = int64(proc.Oid.CreateTS)
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_OID_HPID_INT] = int64(proc.Oid.Hpid)
		if proc.Poid != nil && proc.Poid.UnionType == sfgo.UnionNullOIDTypeEnumOID {
			fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_POID_CREATETS_INT] = proc.Poid.OID.CreateTS
			fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_POID_HPID_INT] = proc.Poid.OID.Hpid
		} else {
			fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_POID_CREATETS_INT] = sfgo.Zeros.Int64
			fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_POID_HPID_INT] = sfgo.Zeros.Int64
		}
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_TS_INT] = proc.Ts
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.PROC_EXE_STR] = proc.Exe
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.PROC_EXEARGS_STR] = proc.ExeArgs
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_UID_INT] = int64(proc.Uid)
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.PROC_USERNAME_STR] = proc.UserName
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_GID_INT] = int64(proc.Gid)
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.PROC_GROUPNAME_STR] = proc.GroupName
		if proc.Tty {
			fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_TTY_INT] = 1
		} else {
			fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_TTY_INT] = 0
		}
		if proc.Entry {
			fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_ENTRY_INT] = 1
		} else {
			fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_ENTRY_INT] = 0
		}
		if proc.ContainerId != nil && proc.ContainerId.UnionType == sfgo.UnionNullStringTypeEnumString {
			fr.Strs[sfgo.SYSFLOW_IDX][sfgo.PROC_CONTAINERID_STRING_STR] = proc.ContainerId.String
		} else {
			fr.Strs[sfgo.SYSFLOW_IDX][sfgo.PROC_CONTAINERID_STRING_STR] = sfgo.Zeros.String
		}
	} else {
		logger.Warn.Println("Event does not have a related process.  This should not happen.")
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_STATE_INT] = sfgo.Zeros.Int64
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_OID_CREATETS_INT] = sfgo.Zeros.Int64
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_OID_HPID_INT] = sfgo.Zeros.Int64
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_POID_CREATETS_INT] = sfgo.Zeros.Int64
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_POID_HPID_INT] = sfgo.Zeros.Int64
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_TS_INT] = sfgo.Zeros.Int64
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.PROC_EXE_STR] = sfgo.Zeros.String
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.PROC_EXEARGS_STR] = sfgo.Zeros.String
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_UID_INT] = sfgo.Zeros.Int64
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.PROC_USERNAME_STR] = sfgo.Zeros.String
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_GID_INT] = sfgo.Zeros.Int64
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.PROC_GROUPNAME_STR] = sfgo.Zeros.String
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_TTY_INT] = sfgo.Zeros.Int64
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.PROC_ENTRY_INT] = sfgo.Zeros.Int64
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.PROC_CONTAINERID_STRING_STR] = sfgo.Zeros.String
	}
	if file != nil {
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FILE_STATE_INT] = int64(file.State)
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FILE_TS_INT] = file.Ts
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FILE_RESTYPE_INT] = int64(file.Restype)
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.FILE_PATH_STR] = file.Path
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.FILE_OID_STR] = getOIDStr(file.Oid[:])
		if file.ContainerId != nil && file.ContainerId.UnionType == sfgo.UnionNullStringTypeEnumString {
			fr.Strs[sfgo.SYSFLOW_IDX][sfgo.FILE_CONTAINERID_STRING_STR] = file.ContainerId.String
		} else {
			fr.Strs[sfgo.SYSFLOW_IDX][sfgo.FILE_CONTAINERID_STRING_STR] = sfgo.Zeros.String
		}
	} else {
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FILE_STATE_INT] = sfgo.Zeros.Int64
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FILE_TS_INT] = sfgo.Zeros.Int64
		fr.Ints[sfgo.SYSFLOW_IDX][sfgo.FILE_RESTYPE_INT] = sfgo.Zeros.Int64
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.FILE_PATH_STR] = sfgo.Zeros.String
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.FILE_CONTAINERID_STRING_STR] = sfgo.Zeros.String
		fr.Strs[sfgo.SYSFLOW_IDX][sfgo.FILE_OID_STR] = sfgo.Zeros.String
	}
}

func getOIDStr(bs []byte) string {
	return hex.EncodeToString(bs)
}

func newFlatRecord() *sfgo.FlatRecord {
	fr := new(sfgo.FlatRecord)
	fr.Sources = make([]sfgo.Source, 1)
	fr.Ints = make([][]int64, 1)
	fr.Strs = make([][]string, 1)
	fr.Sources[sfgo.SYSFLOW_IDX] = sfgo.SYSFLOW_SRC

	fr.Ints[sfgo.SYSFLOW_IDX] = make([]int64, sfgo.INT_ARRAY_SIZE)
	fr.Strs[sfgo.SYSFLOW_IDX] = make([]string, sfgo.STR_ARRAY_SIZE)
	return fr
}
