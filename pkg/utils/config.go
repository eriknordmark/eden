package utils

import (
	"bytes"
	"fmt"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

//ConfigVars struct with parameters from config file
type ConfigVars struct {
	AdamIP         string
	AdamPort       string
	AdamDir        string
	AdamCA         string
	AdamRemote     bool
	EveBaseTag     string
	EveBaseVersion string
	EveHV          string
	SshKey         string
	CheckLogs      bool
	EveCert        string
	EveSerial      string
	ZArch          string
	DevModel       string
	EdenBinDir     string
 	EdenProg       string
	TestProg       string
	TestScript     string
}

//InitVars loads vars from viper
func InitVars() (*ConfigVars, error) {
	configPath, err := DefaultConfigPath()
	if err != nil {
		return nil, err
	}
	loaded, err := LoadConfigFile(configPath)
	if err != nil {
		return nil, err
	}
	if loaded {
		var vars = &ConfigVars{
			AdamIP:         viper.GetString("adam.ip"),
			AdamPort:       viper.GetString("adam.port"),
			AdamDir:        ResolveAbsPath(viper.GetString("adam.dist")),
			AdamCA:         ResolveAbsPath(viper.GetString("adam.ca")),
			SshKey:         ResolveAbsPath(viper.GetString("eden.ssh-key")),
			CheckLogs:      viper.GetBool("eden.logs"),
			EveCert:        ResolveAbsPath(viper.GetString("eve.cert")),
			EveSerial:      viper.GetString("eve.serial"),
			ZArch:          viper.GetString("eve.arch"),
			EveHV:          viper.GetString("eve.hv"),
			EveBaseTag:     viper.GetString("eve.base-tag"),
			EveBaseVersion: fmt.Sprintf("%s-%s-%s", viper.GetString("eve.base-version"), viper.GetString("eve.hv"), viper.GetString("eve.arch")),
			DevModel:       viper.GetString("eve.devmodel"),
			AdamRemote:     viper.GetBool("adam.remote"),
			EdenBinDir:     viper.GetString("eden.bin-dist"),
			EdenProg:       viper.GetString("eden.eden-bin"),
			TestProg:       viper.GetString("eden.test-bin"),
			TestScript:     viper.GetString("eden.test-script"),
		}
		return vars, nil
	}
	return nil, nil
}

//DefaultBaseOSTag for uploadable rootfs
const DefaultBaseOSTag = "4619a12aa4c128972d91539c04938ca3cd0a8ab1"

//DefaultBaseOSVersion for uploadable rootfs
const DefaultBaseOSVersion = "0.0.0-snapshot-master-657e5c1b-2020-05-01.21.53"

var defaultEnvConfig = `#config is generated by eden
adam:
    #tag on adam container to pull
    tag: 0.0.26

    #location of adam
    dist: adam

    #port of adam
    port: 3333

    #domain of adam
    domain: mydomain.adam

    #ip of adam for EVE access
    eve-ip: {{ .EVEIP }}

    #ip of adam for EDEN access
    ip: {{ .IP }}

    #force adam rebuild
    force: true

    #certificate for communication with adam
    ca: adam/run/config/root-certificate.pem

    #use remote adam
    remote: true

    #use v1 api
    v1: true

eve:
    #devmodel
    devmodel: Qemu

    #EVE arch (amd64/arm64)
    arch: {{ .Arch }}

    #EVE os (linux/darwin)
    os: {{ .OS }}

    #EVE acceleration (set to false if you have problems with qemu)
    accel: true

    #variant of hypervisor of EVE (kvm/xen)
    hv: kvm

    #serial number in SMBIOS
    serial: 31415926

    #onboarding certificate of EVE to put into adam
    cert: certs/onboard.cert.pem

    #EVE pid file
    pid: eve.pid

    #EVE log file
    log: eve.log

    #EVE firmware
    firmware: eve/dist/amd64/OVMF.fd

    #eve repo used in clone mode (eden.download = false)
    repo: https://github.com/lf-edge/eve.git

    #eve tag
    tag: 5.3.0

    #eve tag for base os
    base-tag: {{ .DefaultBaseOSTag }}

    #eve version (without hv and os)
    base-version: {{ .DefaultBaseOSVersion }}

    #forward of ports in qemu [(HOST:EVE)]
    hostfwd:
        2222: 22
        5912: 5901
        5911: 5900
        8027: 8027
        8028: 8028

    #location of eve directory
    dist: eve

    #location of EVE base os directory
    base-dist: evebaseos

    #file to save qemu config
    qemu-config: {{ .EdenDir }}/qemu.conf

    #uuid of EVE to use in cert
    uuid: {{ .UUID }}

    #live image of EVE
    image-file: eve/dist/amd64/live.qcow2

    #dtb directory of EVE
    dtb-part: 

    #config part of EVE
    config-part: adam/run/config

eden:
    #root directory of eden
    root: {{ .Root }}
    images:
        #directory to save images
        dist: images

        #yml to build docker image
        docker: {{ .ImageDir }}/docker/alpine/alpine.yml

        #yml to build vm image
        vm: {{ .ImageDir }}/vm/alpine/alpine.yml

    #download eve instead of build
    download: true

    #eserver is tool for serve images
    eserver:
        #ip (domain name) of eserver for EVE access
        ip: mydomain.adam

        #port for eserver
        port: 8888

        #pid for eserver
        pid: eserver.pid

        #log of eserver
        log: eserver.log

    #directory to save certs
    certs-dist: certs

    #directory to save binaries
    bin-dist: bin

    #ssh-key to put into EVE
    ssh-key: certs/id_rsa.pub

    #observe logs in tests
    logs: false

    #eden binary
    eden-bin: eden

    #test binary
    test-bin: eden.integration.test

    #test script
    test-script: eden.integration.tests.txt
`

//DefaultEdenDir returns path to default directory
func DefaultEdenDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, ".eden"), nil
}

//DefaultConfigPath returns path to default config
func DefaultConfigPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, ".eden", "config.yml"), nil
}

//CurrentDirConfigPath returns path to config.yml in current folder
func CurrentDirConfigPath() (string, error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(currentPath, "config.yml"), nil
}

//LoadConfigFile load config from file with viper
func LoadConfigFile(config string) (loaded bool, err error) {
	if config == "" {
		config, err = DefaultConfigPath()
		if err != nil {
			return false, fmt.Errorf("fail in DefaultConfigPath: %s", err.Error())
		}
	}
	if _, err = os.Stat(config); os.IsNotExist(err) {
		if err = GenerateConfigFile(config); err != nil {
			return false, fmt.Errorf("fail in generate yaml: %s", err.Error())
		} else {
			log.Infof("Config file generated: %s", config)
		}
	}
	abs, err := filepath.Abs(config)
	if err != nil {
		return false, fmt.Errorf("fail in reading filepath: %s", err.Error())
	}
	viper.SetConfigFile(abs)
	if err := viper.ReadInConfig(); err != nil {
		return false, fmt.Errorf("failed to read config file: %s", err.Error())
	}
	currentFolderDir, err := CurrentDirConfigPath()
	if err != nil {
		log.Errorf("CurrentDirConfigPath: %s", err)
	} else {
		log.Debugf("Try to add config from %s", currentFolderDir)
		if _, err = os.Stat(currentFolderDir); !os.IsNotExist(err) {
			abs, err = filepath.Abs(currentFolderDir)
			if err != nil {
				log.Errorf("CurrentDirConfigPath absolute: %s", err)
			} else {
				viper.SetConfigFile(abs)
				if err := viper.MergeInConfig(); err != nil {
					log.Errorf("failed in merge config file: %s", err.Error())
				} else {
					log.Debugf("Merged config with %s", abs)
				}
			}
		}
	}
	return true, nil
}

//GenerateConfigFile is a function to generate default yml
func GenerateConfigFile(filePath string) error {
	currentPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		log.Fatal(err)
	}
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	edenDir, err := DefaultEdenDir()
	if err != nil {
		log.Fatal(err)
	}

	t := template.New("t")
	_, err = t.Parse(defaultEnvConfig)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	ip, err := GetIPForDockerAccess()
	if err != nil {
		return err
	}
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	nets, err := GetSubnetsNotUsed(1)
	if err != nil {
		return err
	}
	address := strings.Split(nets[0].FirstAddress.String(), ".")
	eveIP := strings.Join(append(strings.Split(nets[0].FirstAddress.String(), ".")[:len(address)-1], "2"), ".")
	err = t.Execute(buf,
		struct {
			ImageDir             string
			Root                 string
			IP                   string
			EVEIP                string
			UUID                 string
			Arch                 string
			OS                   string
			EdenDir              string
			DefaultBaseOSVersion string
			DefaultBaseOSTag     string
		}{
			ImageDir:             filepath.Join(currentPath, "images"),
			Root:                 filepath.Join(currentPath, "dist"),
			IP:                   ip,
			EVEIP:                eveIP,
			UUID:                 id.String(),
			Arch:                 runtime.GOARCH,
			OS:                   runtime.GOOS,
			EdenDir:              edenDir,
			DefaultBaseOSVersion: DefaultBaseOSVersion,
			DefaultBaseOSTag:     DefaultBaseOSTag,
		})
	if err != nil {
		return err
	}
	_, err = file.Write(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
