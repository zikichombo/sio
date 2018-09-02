package entry

var names = [...]string{"CoreAudio Audio Queue Services", "CoreAudio AUHAL", "CoreAudio RemoteIO"}

func Names() []string {
	res := make([]string, len(names))
	copy(res, names)
	return res
}
