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

package spec

import (
	"os"
	"testing"

	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestDefaultDataDir(t *testing.T) {
	// Test with without global DataDir.
	topo := new(Specification)
	topo.Masters = append(topo.Masters, &MasterSpec{Host: "1.1.1.1", Port: 1111})
	topo.Workers = append(topo.Workers, &WorkerSpec{Host: "1.1.2.1", Port: 2221})
	data, err := yaml.Marshal(topo)
	require.NoError(t, err)

	// Check default value.
	topo = new(Specification)
	err = yaml.Unmarshal(data, topo)
	require.NoError(t, err)
	require.Equal(t, "data", topo.GlobalOptions.DataDir)
	require.Equal(t, "data", topo.Masters[0].DataDir)
	require.Equal(t, "data", topo.Workers[0].DataDir)

	// Can keep the default value.
	data, err = yaml.Marshal(topo)
	require.NoError(t, err)
	topo = new(Specification)
	err = yaml.Unmarshal(data, topo)
	require.NoError(t, err)
	require.Equal(t, "data", topo.GlobalOptions.DataDir)
	require.Equal(t, "data", topo.Masters[0].DataDir)
	require.Equal(t, "data", topo.Workers[0].DataDir)

	// Test with global DataDir.
	topo = new(Specification)
	topo.GlobalOptions.DataDir = "/gloable_data"
	topo.Masters = append(topo.Masters, &MasterSpec{Host: "1.1.1.1", Port: 1111})
	topo.Masters = append(topo.Masters, &MasterSpec{Host: "1.1.1.2", Port: 1112, DataDir: "/my_data"})
	topo.Workers = append(topo.Workers, &WorkerSpec{Host: "1.1.2.1", Port: 2221})
	topo.Workers = append(topo.Workers, &WorkerSpec{Host: "1.1.2.2", Port: 2222, DataDir: "/my_data"})
	data, err = yaml.Marshal(topo)
	require.NoError(t, err)

	topo = new(Specification)
	err = yaml.Unmarshal(data, topo)
	require.NoError(t, err)
	require.Equal(t, "/gloable_data", topo.GlobalOptions.DataDir)
	require.Equal(t, "/gloable_data/dm-master-1111", topo.Masters[0].DataDir)
	require.Equal(t, "/my_data", topo.Masters[1].DataDir)
	require.Equal(t, "/gloable_data/dm-worker-2221", topo.Workers[0].DataDir)
	require.Equal(t, "/my_data", topo.Workers[1].DataDir)
}

func TestGlobalOptions(t *testing.T) {
	topo := Specification{}
	err := yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "test-deploy"
  data_dir: "test-data"
master_servers:
  - host: 172.16.5.138
    deploy_dir: "master-deploy"
worker_servers:
  - host: 172.16.5.53
    data_dir: "worker-data"
`), &topo)
	require.NoError(t, err)
	require.Equal(t, "test1", topo.GlobalOptions.User)

	require.Equal(t, 220, topo.GlobalOptions.SSHPort)
	require.Equal(t, 220, topo.Masters[0].SSHPort)
	require.Equal(t, "master-deploy", topo.Masters[0].DeployDir)

	require.Equal(t, 220, topo.Workers[0].SSHPort)
	require.Equal(t, "test-deploy/dm-worker-8262", topo.Workers[0].DeployDir)
	require.Equal(t, "worker-data", topo.Workers[0].DataDir)
}

func TestDirectoryConflicts(t *testing.T) {
	topo := Specification{}
	err := yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "test-deploy"
  data_dir: "test-data"
master_servers:
  - host: 172.16.5.138
    deploy_dir: "/test-1"
worker_servers:
  - host: 172.16.5.138
    data_dir: "/test-1"
`), &topo)
	require.Error(t, err)
	require.Equal(t, "directory conflict for '/test-1' between 'master_servers:172.16.5.138.deploy_dir' and 'worker_servers:172.16.5.138.data_dir'", err.Error())

	err = yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "test-deploy"
  data_dir: "/test-data"
master_servers:
  - host: 172.16.5.138
    data_dir: "test-1"
worker_servers:
  - host: 172.16.5.138
    data_dir: "test-1"
`), &topo)
	require.NoError(t, err)
}

func TestPortConflicts(t *testing.T) {
	topo := Specification{}
	err := yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "test-deploy"
  data_dir: "test-data"
master_servers:
  - host: 172.16.5.138
    peer_port: 1234
worker_servers:
  - host: 172.16.5.138
    port: 1234
`), &topo)
	require.Error(t, err)
	require.Equal(t, "port conflict for '1234' between 'master_servers:172.16.5.138.peer_port,omitempty' and 'worker_servers:172.16.5.138.port,omitempty'", err.Error())
}

func TestPlatformConflicts(t *testing.T) {
	// aarch64 and arm64 are equal
	topo := Specification{}
	err := yaml.Unmarshal([]byte(`
global:
  os: "linux"
  arch: "aarch64"
master_servers:
  - host: 172.16.5.138
    arch: "arm64"
worker_servers:
  - host: 172.16.5.138
    arch: "aarch64"
`), &topo)
	require.NoError(t, err)

	// different arch defined for the same host
	topo = Specification{}
	err = yaml.Unmarshal([]byte(`
global:
  os: "linux"
master_servers:
  - host: 172.16.5.138
    arch: "aarch64"
    os: "linux"
worker_servers:
  - host: 172.16.5.138
    arch: "amd64"
    os: "linux"
`), &topo)
	require.Error(t, err)
	require.Equal(t, "platform mismatch for '172.16.5.138' between 'master_servers:linux/arm64' and 'worker_servers:linux/amd64'", err.Error())

	// different os defined for the same host
	topo = Specification{}
	err = yaml.Unmarshal([]byte(`
global:
  os: "linux"
  arch: "aarch64"
master_servers:
  - host: 172.16.5.138
    os: "darwin"
    arch: "aarch64"
worker_servers:
  - host: 172.16.5.138
    os: "linux"
    arch: "aarch64"
`), &topo)
	require.Error(t, err)
	require.Equal(t, "platform mismatch for '172.16.5.138' between 'master_servers:darwin/arm64' and 'worker_servers:linux/arm64'", err.Error())
}

func TestCountDir(t *testing.T) {
	topo := Specification{}

	err := yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "test-deploy"
master_servers:
  - host: 172.16.5.138
    deploy_dir: "master-deploy"
    data_dir: "/test-data/data-1"
worker_servers:
  - host: 172.16.5.53
    data_dir: "test-1"
`), &topo)
	require.NoError(t, err)
	cnt := topo.CountDir("172.16.5.53", "test-deploy/dm-worker-8262")
	require.Equal(t, 3, cnt)
	cnt = topo.CountDir("172.16.5.138", "/test-data/data")
	require.Equal(t, 0, cnt) // should not match partial path

	err = yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "/test-deploy"
master_servers:
  - host: 172.16.5.138
    deploy_dir: "master-deploy"
    data_dir: "/test-data/data-1"
worker_servers:
  - host: 172.16.5.138
    data_dir: "/test-data/data-2"
`), &topo)
	require.NoError(t, err)
	cnt = topo.CountDir("172.16.5.138", "/test-deploy/dm-worker-8262")
	require.Equal(t, 2, cnt)
	cnt = topo.CountDir("172.16.5.138", "")
	require.Equal(t, 2, cnt)
	cnt = topo.CountDir("172.16.5.138", "test-data")
	require.Equal(t, 0, cnt)
	cnt = topo.CountDir("172.16.5.138", "/test-data")
	require.Equal(t, 2, cnt)

	err = yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "/test-deploy"
  data_dir: "/test-data"
master_servers:
  - host: 172.16.5.138
    data_dir: "data-1"
worker_servers:
  - host: 172.16.5.138
    data_dir: "data-2"
  - host: 172.16.5.53
`), &topo)
	require.NoError(t, err)
	// if per-instance data_dir is set, the global data_dir is ignored, and if it
	// is a relative path, it will be under the instance's deploy_dir
	cnt = topo.CountDir("172.16.5.138", "/test-deploy/dm-worker-8262")
	require.Equal(t, 3, cnt)
	cnt = topo.CountDir("172.16.5.138", "")
	require.Equal(t, 0, cnt)
	cnt = topo.CountDir("172.16.5.53", "/test-data")
	require.Equal(t, 1, cnt)
}

func withTempFile(t *testing.T, content string, fn func(string)) {
	file, err := os.CreateTemp("/tmp", "topology-test")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	_, err = file.WriteString(content)
	require.NoError(t, err)
	file.Close()

	fn(file.Name())
}

func with2TempFile(t *testing.T, content1, content2 string, fn func(string, string)) {
	withTempFile(t, content1, func(file1 string) {
		withTempFile(t, content2, func(file2 string) {
			fn(file1, file2)
		})
	})
}

func merge4test(base, scale string) (*Specification, error) {
	baseTopo := Specification{}
	if err := spec.ParseTopologyYaml(base, &baseTopo); err != nil {
		return nil, err
	}

	scaleTopo := baseTopo.NewPart()
	if err := spec.ParseTopologyYaml(scale, scaleTopo); err != nil {
		return nil, err
	}

	mergedTopo := baseTopo.MergeTopo(scaleTopo)
	if err := mergedTopo.Validate(); err != nil {
		return nil, err
	}

	return mergedTopo.(*Specification), nil
}

func TestRelativePath(t *testing.T) {
	// base test
	withTempFile(t, `
master_servers:
  - host: 172.16.5.140
worker_servers:
  - host: 172.16.5.140
`, func(file string) {
		topo := Specification{}
		err := spec.ParseTopologyYaml(file, &topo)
		require.NoError(t, err)
		spec.ExpandRelativeDir(&topo)
		require.Equal(t, "/home/tidb/deploy/dm-master-8261", topo.Masters[0].DeployDir)
		require.Equal(t, "/home/tidb/deploy/dm-worker-8262", topo.Workers[0].DeployDir)
	})

	// test data dir & log dir
	withTempFile(t, `
master_servers:
  - host: 172.16.5.140
    deploy_dir: my-deploy
    data_dir: my-data
    log_dir: my-log
`, func(file string) {
		topo := Specification{}
		err := spec.ParseTopologyYaml(file, &topo)
		require.NoError(t, err)
		spec.ExpandRelativeDir(&topo)

		require.Equal(t, "/home/tidb/my-deploy", topo.Masters[0].DeployDir)
		require.Equal(t, "/home/tidb/my-deploy/my-data", topo.Masters[0].DataDir)
		require.Equal(t, "/home/tidb/my-deploy/my-log", topo.Masters[0].LogDir)
	})

	// test global options, case 1
	withTempFile(t, `
global:
  deploy_dir: my-deploy
master_servers:
  - host: 172.16.5.140
`, func(file string) {
		topo := Specification{}
		err := spec.ParseTopologyYaml(file, &topo)
		require.NoError(t, err)
		spec.ExpandRelativeDir(&topo)

		require.Equal(t, "/home/tidb/my-deploy/dm-master-8261", topo.Masters[0].DeployDir)
		require.Equal(t, "/home/tidb/my-deploy/dm-master-8261/data", topo.Masters[0].DataDir)
		require.Equal(t, "", topo.Masters[0].LogDir)
	})

	// test global options, case 2
	withTempFile(t, `
global:
  deploy_dir: my-deploy
master_servers:
  - host: 172.16.5.140
worker_servers:
  - host: 172.16.5.140
    port: 20160
  - host: 172.16.5.140
    port: 20161
`, func(file string) {
		topo := Specification{}
		err := spec.ParseTopologyYaml(file, &topo)
		require.NoError(t, err)
		spec.ExpandRelativeDir(&topo)

		require.Equal(t, "my-deploy", topo.GlobalOptions.DeployDir)
		require.Equal(t, "data", topo.GlobalOptions.DataDir)

		require.Equal(t, "/home/tidb/my-deploy/dm-worker-20160", topo.Workers[0].DeployDir)
		require.Equal(t, "/home/tidb/my-deploy/dm-worker-20160/data", topo.Workers[0].DataDir)

		require.Equal(t, "/home/tidb/my-deploy/dm-worker-20161", topo.Workers[1].DeployDir)
		require.Equal(t, "/home/tidb/my-deploy/dm-worker-20161/data", topo.Workers[1].DataDir)
	})

	// test global options, case 3
	withTempFile(t, `
global:
  deploy_dir: my-deploy
master_servers:
  - host: 172.16.5.140
worker_servers:
  - host: 172.16.5.140
    port: 20160
    data_dir: my-data
    log_dir: my-log
  - host: 172.16.5.140
    port: 20161
`, func(file string) {
		topo := Specification{}
		err := spec.ParseTopologyYaml(file, &topo)
		require.NoError(t, err)
		spec.ExpandRelativeDir(&topo)

		require.Equal(t, "my-deploy", topo.GlobalOptions.DeployDir)
		require.Equal(t, "data", topo.GlobalOptions.DataDir)

		require.Equal(t, "/home/tidb/my-deploy/dm-worker-20160", topo.Workers[0].DeployDir)
		require.Equal(t, "/home/tidb/my-deploy/dm-worker-20160/my-data", topo.Workers[0].DataDir)
		require.Equal(t, "/home/tidb/my-deploy/dm-worker-20160/my-log", topo.Workers[0].LogDir)

		require.Equal(t, "/home/tidb/my-deploy/dm-worker-20161", topo.Workers[1].DeployDir)
		require.Equal(t, "/home/tidb/my-deploy/dm-worker-20161/data", topo.Workers[1].DataDir)
		require.Equal(t, "", topo.Workers[1].LogDir)
	})

	// test global options, case 4
	withTempFile(t, `
global:
  data_dir: my-global-data
  log_dir: my-global-log
master_servers:
  - host: 172.16.5.140
worker_servers:
  - host: 172.16.5.140
    port: 20160
    data_dir: my-local-data
    log_dir: my-local-log
  - host: 172.16.5.140
    port: 20161
`, func(file string) {
		topo := Specification{}
		err := spec.ParseTopologyYaml(file, &topo)
		require.NoError(t, err)
		spec.ExpandRelativeDir(&topo)

		require.Equal(t, "deploy", topo.GlobalOptions.DeployDir)
		require.Equal(t, "my-global-data", topo.GlobalOptions.DataDir)
		require.Equal(t, "my-global-log", topo.GlobalOptions.LogDir)

		require.Equal(t, "/home/tidb/deploy/dm-worker-20160", topo.Workers[0].DeployDir)
		require.Equal(t, "/home/tidb/deploy/dm-worker-20160/my-local-data", topo.Workers[0].DataDir)
		require.Equal(t, "/home/tidb/deploy/dm-worker-20160/my-local-log", topo.Workers[0].LogDir)

		require.Equal(t, "/home/tidb/deploy/dm-worker-20161", topo.Workers[1].DeployDir)
		require.Equal(t, "/home/tidb/deploy/dm-worker-20161/my-global-data", topo.Workers[1].DataDir)
		require.Equal(t, "/home/tidb/deploy/dm-worker-20161/my-global-log", topo.Workers[1].LogDir)
	})
}

func TestTopologyMerge(t *testing.T) {
	// base test
	with2TempFile(t, `
master_servers:
  - host: 172.16.5.140
worker_servers:
  - host: 172.16.5.140
`, `
worker_servers:
  - host: 172.16.5.139
`, func(base, scale string) {
		topo, err := merge4test(base, scale)
		require.NoError(t, err)
		spec.ExpandRelativeDir(topo)

		require.Equal(t, "/home/tidb/deploy/dm-worker-8262", topo.Workers[0].DeployDir)
		require.Equal(t, "/home/tidb/deploy/dm-worker-8262/data", topo.Workers[0].DataDir)
		require.Equal(t, "", topo.Workers[0].LogDir)

		require.Equal(t, "/home/tidb/deploy/dm-worker-8262", topo.Workers[1].DeployDir)
		require.Equal(t, "/home/tidb/deploy/dm-worker-8262/data", topo.Workers[1].DataDir)
		require.Equal(t, "", topo.Workers[1].LogDir)
	})

	// test global option overwrite
	with2TempFile(t, `
global:
  user: test
  deploy_dir: /my-global-deploy
master_servers:
  - host: 172.16.5.140
worker_servers:
  - host: 172.16.5.140
    log_dir: my-local-log
    data_dir: my-local-data
  - host: 172.16.5.175
    deploy_dir: flash-deploy
  - host: 172.16.5.141
`, `
worker_servers:
  - host: 172.16.5.139
    deploy_dir: flash-deploy
  - host: 172.16.5.134
`, func(base, scale string) {
		topo, err := merge4test(base, scale)
		require.NoError(t, err)

		spec.ExpandRelativeDir(topo)

		require.Equal(t, "/my-global-deploy/dm-worker-8262", topo.Workers[0].DeployDir)
		require.Equal(t, "/my-global-deploy/dm-worker-8262/my-local-data", topo.Workers[0].DataDir)
		require.Equal(t, "/my-global-deploy/dm-worker-8262/my-local-log", topo.Workers[0].LogDir)

		require.Equal(t, "/home/test/flash-deploy", topo.Workers[1].DeployDir)
		require.Equal(t, "/home/test/flash-deploy/data", topo.Workers[1].DataDir)
		require.Equal(t, "/home/test/flash-deploy", topo.Workers[3].DeployDir)
		require.Equal(t, "/home/test/flash-deploy/data", topo.Workers[3].DataDir)

		require.Equal(t, "/my-global-deploy/dm-worker-8262", topo.Workers[2].DeployDir)
		require.Equal(t, "/my-global-deploy/dm-worker-8262/data", topo.Workers[2].DataDir)
		require.Equal(t, "/my-global-deploy/dm-worker-8262", topo.Workers[4].DeployDir)
		require.Equal(t, "/my-global-deploy/dm-worker-8262/data", topo.Workers[4].DataDir)
	})
}

func TestMonitorLogDir(t *testing.T) {
	withTempFile(t, `
monitored:
  node_exporter_port: 39100
  blackbox_exporter_port: 39115
  deploy_dir: "test-deploy"
  log_dir: "test-deploy/log"
`, func(file string) {
		topo := Specification{}
		err := spec.ParseTopologyYaml(file, &topo)
		require.NoError(t, err)
		require.Equal(t, 39100, topo.MonitoredOptions.NodeExporterPort)
		require.Equal(t, 39115, topo.MonitoredOptions.BlackboxExporterPort)
		require.Equal(t, "test-deploy/log", topo.MonitoredOptions.LogDir)
		require.Equal(t, "test-deploy", topo.MonitoredOptions.DeployDir)
	})
}
