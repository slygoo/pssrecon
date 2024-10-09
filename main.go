package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/jfjallid/go-smb/smb"
	smbdcerpc "github.com/jfjallid/go-smb/smb/dcerpc"
	"github.com/jfjallid/go-smb/smb/dcerpc/msrrp"
	"github.com/jfjallid/go-smb/spnego"
)

var sccmlocation = `SOFTWARE\Microsoft\SMS`

func main() {
	Enumerate()
}

func Enumerate() error {
	host := flag.String("host", "", "IP of the sccm PSS or DP server")
	u := flag.String("u", "", "Username to authenticate with")
	p := flag.String("p", "", "Password to authenticate with")
	d := flag.String("d", "", "Domain FQDN to authenticate with")
	hash := flag.String("hash", "", "NTLM Hash to authenticate with")
	flag.Parse()
	var err error
	var options smb.Options
	if *hash != "" {
		var hashBytes = []byte{}
		hashBytes, err = hex.DecodeString(*hash)
		if err != nil {
			return err
		}
		options = smb.Options{
			Host: *host,
			Port: 445,
			Initiator: &spnego.NTLMInitiator{
				User:   *u,
				Domain: *d,
				Hash:   hashBytes,
			},
		}
	} else {
		options = smb.Options{
			Host: *host,
			Port: 445,
			Initiator: &spnego.NTLMInitiator{
				User:     *u,
				Password: *p,
				Domain:   *d,
			},
		}
	}
	session, err := smb.NewConnection(options)
	if err != nil {
		return err
	}
	share := "IPC$"
	err = session.TreeConnect(share)
	if err != nil {
		return err
	}
	f, err := session.OpenFile(share, "winreg")
	if err != nil {
		//sleeping here cuz sometimes after i req it it will open then i can connect
		time.Sleep(1 * time.Second)
		f, err = session.OpenFile(share, "winreg")
		if err != nil {
			return err
		}
	}
	defer f.Close()
	bind, err := smbdcerpc.Bind(f, msrrp.MSRRPUuid, msrrp.MSRRPMajorVersion, msrrp.MSRRPMinorVersion, msrrp.NDRUuid)
	if err != nil {
		return err
	}
	reg := msrrp.NewRPCCon(bind)
	err = EnumeratePSSRoles(reg)
	if err != nil {
		return err
	}
	err = EnumerateSiteDB(reg)
	if err != nil {
		return errors.New("[-] Could Not Locate Site DB")
	}
	return nil
}

func EnumerateSiteDB(reg *msrrp.RPCCon) error {

	hKey, err := reg.OpenBaseKey(msrrp.HKEYLocalMachine)
	if err != nil {
		return err
	}
	subkey := sccmlocation + `\COMPONENTS\SMS_SITE_COMPONENT_MANAGER\Multisite Component Servers`
	listedSubKeys, err := reg.GetSubKeyNames(hKey, subkey)
	if err != nil {
		return err
	}
	if len(listedSubKeys) == 1 {
		fmt.Println("[+] Site Database Found: " + listedSubKeys[0])
	}
	if len(listedSubKeys) == 0 {
		fmt.Println("[+] Site Database is Local to the Primary Site Server")
	}
	if len(listedSubKeys) > 1 {
		for i := 0; i < len(listedSubKeys); i++ {
			fmt.Println("[+] Multisite Component Server: " + listedSubKeys[i])
		}
	}
	return nil
}

func EnumeratePSSRoles(reg *msrrp.RPCCon) error {
	hKey, err := reg.OpenBaseKey(msrrp.HKEYLocalMachine)
	if err != nil {
		return err
	}
	listedSubKeys, err := reg.GetSubKeyNames(hKey, sccmlocation)
	if err != nil {
		return err
	}
	for i := 0; i < len(listedSubKeys); i++ {
		if listedSubKeys[i] == "DP" {
			fmt.Println("[+] Distrubution Point Installed")
			EnumerateDP(reg)
		}
		if listedSubKeys[i] == "MP" {
			fmt.Println("[+] Management Point Installed")
		}
	}
	return nil
}

func EnumerateDP(reg *msrrp.RPCCon) error {
	hKey, err := reg.OpenBaseKey(msrrp.HKEYLocalMachine)
	if err != nil {
		return err
	}
	subkey := sccmlocation + `\DP`
	hSubKey, err := reg.OpenSubKey(hKey, subkey)
	if err != nil {
		return err
	}
	sc, err := reg.QueryValueString(hSubKey, "SiteCode")
	if err != nil {
		return err
	}
	fmt.Println("[+] Site Code Found: " + sc)
	ss, err := reg.QueryValueString(hSubKey, "SiteServer")
	if err != nil {
		return err
	}
	fmt.Println("[+] Site Server Found: " + ss)
	mps, err := reg.QueryValueString(hSubKey, "ManagementPoints")
	if err != nil {
		return err
	}
	split := strings.Split(mps, "*")
	for i := 0; i < len(split); i++ {
		fmt.Println("[+] Management Point Found: " + split[i])
	}
	isAnonymousAccessEnabled, err := reg.QueryValue(hSubKey, "IsAnonymousAccessEnabled")
	if err != nil {
		return err
	}
	if binary.LittleEndian.Uint32(isAnonymousAccessEnabled) == 1 {
		fmt.Println("[+] Anonymous Access On This Distrubution Point Is Enabled")
	}
	isPXE, err := reg.QueryValue(hSubKey, "IsPXE")
	if err != nil {
		return err
	}
	if binary.LittleEndian.Uint32(isPXE) == 1 {
		fmt.Println("[+] PXE Installed")
	}
	return nil
}
