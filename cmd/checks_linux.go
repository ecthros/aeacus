package cmd

import (
	"os/exec"
	"strings"
)

// processCheck (Linux) will process Linux-specific checks
// handed to it by the processCheckWrapper function
func processCheck(check *check, checkType, arg1, arg2, arg3 string) bool {
	switch checkType {
	case "GuestDisabledLDM":
		if check.Message == "" {
			check.Message = "Guest is disabled"
		}
		result, err := guestDisabledLDM()
		return err == nil && result
	case "GuestDisabledLDMNot":
		if check.Message == "" {
			check.Message = "Guest is enabled"
		}
		result, err := guestDisabledLDM()
		return err == nil && !result
	case "PasswordChanged":
		if check.Message == "" {
			check.Message = "Password for " + arg1 + " has been changed"
		}
		result, err := passwordChanged(arg1, arg2)
		return err == nil && result
	case "PasswordChangedNot":
		if check.Message == "" {
			check.Message = "Password for " + arg1 + " has not been changed"
		}
		result, err := passwordChanged(arg1, arg2)
		return err == nil && !result
	case "KernelVersion":
		if check.Message == "" {
			check.Message = "Kernel is version " + arg1
		}
		result, err := kernelVersion(arg1)
		return err == nil && result
	case "KernelVersionNot":
		if check.Message == "" {
			check.Message = "Kernel is not version " + arg1
		}
		result, err := kernelVersion(arg1)
		return err == nil && !result
	case "AutoCheckUpdatesEnabled":
		if check.Message == "" {
			check.Message = "The system automatically checks for updates daily"
		}
		result, err := autoCheckUpdatesEnabled()
		return err == nil && result
	case "AutoCheckUpdatesEnabledNot":
		if check.Message == "" {
			check.Message = "The system does not automatically checks for updates daily"
		}
		result, err := autoCheckUpdatesEnabled()
		return err == nil && !result
	case "PermissionIs":
		if check.Message == "" {

			if arg2 == "octal" {
				check.Message = "The octal permissions of " + arg1 + " are " + arg3
			} else if arg2 == "WorldWritable" {
				check.Message = arg1 + " is world writable"
			} else if arg2 == "WorldReadable" {
				check.Message = arg1 + " is world readable"
			} else {
				check.Message = "Permissions of " + arg1 + " are " + arg3
			}

		}
		result, err := permissionIs(arg1, arg2, arg3)
		return err == nil && result
	case "PermissionIsNot":
		if check.Message == "" {

			if arg2 == "octal" {
				check.Message = "The octal permissions of " + arg1 + " are not " + arg3
			} else if arg2 == "WorldWritable" {
				check.Message = arg1 + " is not world writable"
			} else if arg2 == "WorldReadable" {
				check.Message = arg1 + " is not world readable"
			} else {
				check.Message = "Permissions of " + arg1 + " are not " + arg3
			}

		}
		result, err := permissionIs(arg1, arg2, arg3)
		return err == nil && !result
	default:
		failPrint("No check type " + checkType)
	}
	return false
}

func command(commandGiven string) (bool, error) {
	cmd := rawCmd(commandGiven)
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func commandInterface(progName string, c ...string) (bool, error) {
	cmd := exec.Command(progName, c...)
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func programInstalled(programName string) (bool, error) {
	return commandInterface("/usr/bin/dpkg", "-s", programName)
}

func serviceUp(serviceName string) (bool, error) {
	// TODO: detect and use other init systems
	ret, err := commandContains("systemctl is-active "+serviceName, "inactive")
	return !ret, err
}

func userExists(userName string) (bool, error) {
	return fileContains("/etc/passwd", userName+":x:")
}

func userInGroup(userName, groupName string) (bool, error) {
	return commandContains("groups "+userName, groupName)
}

func firewallUp() (bool, error) {
	status, err := commandOutput("ufw status")
	return status == "Status: active", err
}

func passwordChanged(user, hash string) (bool, error) {
	res, err := fileContains("/etc/shadow", hash)
	return !res, err
}

func guestDisabledLDM() (bool, error) {
	result, err := dirContainsRegex("/usr/share/lightdm/lightdm.conf.d/", "allow-guest( |)=( |)false")
	if !result && err == nil {
		result, err = dirContainsRegex("/etc/lightdm/", "allow-guest( |)=( |)false")
	}
	return result, err
}

func programVersion(programName, versionNum string) (bool, error) {
	commandGiven := `dpkg -l | awk '$2=="` + programName + `" { print $3 }'`
	out, err := rawCmd(commandGiven).Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == versionNum, nil
}

func kernelVersion(version string) (bool, error) {
	return commandContains("uname -r", version)
}

func autoCheckUpdatesEnabled() (bool, error) {
	return dirContainsRegex("/etc/apt/apt.conf.d/", `APT::Periodic::Update-Package-Lists\s+"1";`)
}

// func permissionIs(filePath, permissionToCheck string) (bool, error) {
// 	perm, err := commandOutput(`stat -c '%a' ` + filePath)
// 	return perm == permissionToCheck, err
// }

func permissionIs(filePath, checkType, permissionToCheck string) (bool, error) {
	if checkType == "octal" {
		perm, err := commandOutput(`stat -c '%a' ` + filePath)
		return perm == permissionToCheck, err
	} else if checkType == "WorldWritable" {
		return commandContains(`find `+filePath+` -perm -g+w -or -perm -o+w`, filePath)
	} else if checkType == "WorldReadable" {
		return commandContains(`find `+filePath+` -perm -o=r`, filePath)
	}
	// If arguments are messed up or whatever:
	return false, nil
}
