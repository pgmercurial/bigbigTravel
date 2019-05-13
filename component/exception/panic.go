package exception

func Panic(logicErr interface{}, sysErr error)  {
	pe := &PanicError{
		LogicErr: logicErr,
		SysErr:   sysErr,
	}
	panic(pe)
}

