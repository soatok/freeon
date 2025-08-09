package internal

func InitKeyGenCeremony(host string, participants int, threshold int) {

}

func JoinKeyGenCeremony(host, groupID, recpient string) {

}

func ListKeyGen() {

}

func InitSignCeremony(groupID, host string, message []byte, openssh bool) {

}

func JoinSignCeremony(ceremonyID, host, identityFile string, message []byte, autoConfirm bool) {

}

func ListSign() {

}

func TerminateSignCeremony(ceremonyID string) {

}
