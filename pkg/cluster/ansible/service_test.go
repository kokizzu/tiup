// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package ansible

import (
	"testing"

	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/stretchr/testify/require"
)

var tiflashConfig = `
default_profile = "default"
display_name = "TiFlash"
listen_host = "0.0.0.0"
path = "/data1/test-cluster/leiysky-ansible-test-deploy/tiflash/data/db"
tmp_path = "/data1/test-cluster/leiysky-ansible-test-deploy/tiflash/data/db/tmp"

[flash]
service_addr = "172.16.5.85:11317"
tidb_status_addr = "172.16.5.85:11310"
[flash.flash_cluster]
cluster_manager_path = "/data1/test-cluster/leiysky-ansible-test-deploy/bin/tiflash/flash_cluster_manager"
log = "/data1/test-cluster/leiysky-ansible-test-deploy/log/tiflash_cluster_manager.log"
master_ttl = 60
refresh_interval = 20
update_rule_interval = 5
[flash.proxy]
config = "/data1/test-cluster/leiysky-ansible-test-deploy/conf/tiflash-learner.toml"
`

func TestParseTiflashConfigFromFileData(t *testing.T) {
	specObj := new(spec.TiFlashSpec)
	data := []byte(tiflashConfig)

	err := parseTiflashConfigFromFileData(specObj, data)
	require.NoError(t, err)

	require.Equal(t, "/data1/test-cluster/leiysky-ansible-test-deploy/tiflash/data/db", specObj.DataDir)
	require.Equal(t, "/data1/test-cluster/leiysky-ansible-test-deploy/tiflash/data/db/tmp", specObj.TmpDir)
}
