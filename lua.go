package cron

import (
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/xcall"
	"github.com/rock-go/rock/logger"
	"reflect"
)

var CRON = reflect.TypeOf((*Cron)(nil)).String()

func (c *Cron)LAddFunc(L *lua.LState) int {
	n := L.GetTop()
	if n != 3 {
		L.RaiseError("invalid args , usage add(string , function , label)")
		return 0
	}

	spec := L.CheckString(1)
	fn := L.CheckFunction(2)
	label := L.CheckString(3)

	eid , err := c.AddFunc(spec , func(){
		//这里注意 多个函数同时触发
		co := lua.State()
		co.A = L.A

		e := xcall.CallByEnv(co , fn , xcall.Rock)
		if e != nil {
			logger.Errorf("%v" , e)
		}
		//回收co虚拟机
		lua.FreeState(co)
	})

	if err != nil {
		L.RaiseError("%v" , err)
		return 0
	}

	c.masks = append(c.masks , newMask(spec , label))

	L.Push(lua.LNumber(eid))
	return 1
}


func (c *Cron) Index(L *lua.LState , key string) lua.LValue {
	if key == "add_func" { return lua.NewFunction(c.LAddFunc)}
	return lua.LNil
}

func newLuaCron(L *lua.LState) int {
	name := L.CheckString(1)
	proc := L.NewProc(name , CRON)
	proc.Set(New(name))
	L.Push(proc)
	return 1
}

func LuaInjectApi( env xcall.Env) {
	env.Set("cron" , lua.NewFunction(newLuaCron))
}