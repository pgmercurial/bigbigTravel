package exception

const (
	EmojiHappy      = 1
	EmojiCry        = 2
	EmojiDaze       = 3
	EmojiCoverMouth = 4
)

const (
	ErrorLevelNormal  = 1
	ErrorLevelWarning = 2
	ErrorLevelAnomaly = 3
	ErrorLevelError   = 4
)

type RunException struct {
	Code       int
	DisplayMsg string
	LogMsg     string
	Emoji      int
	ErrorLevel int
}

type RunError struct {
	Code       int
	DisplayMsg string
	LogMsg     string
}

type PanicError struct {
	LogicErr interface{}
	SysErr   error
}

func DefineException(code int, displayMsg string, logMsg string, errorLevel ...int) *RunException {
	e := &RunException{
		Code:       code,
		DisplayMsg: displayMsg,
		LogMsg:     logMsg,
	}
	if len(errorLevel) > 0 {
		e.ErrorLevel = errorLevel[0]
	}
	return e
}

func DefineError(code int, displayMsg string, logMsg string) *RunError {
	e := &RunError{
		Code:       code,
		DisplayMsg: displayMsg,
		LogMsg:     logMsg,
	}
	return e
}

func EmojiException(code int, displayMsg string, logMsg string, emoji int, errorLevel ...int) *RunException {
	e := DefineException(code, displayMsg, logMsg, errorLevel...)
	e.Emoji = emoji
	return e
}
