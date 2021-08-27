# rock-cron-go
# 磐石系统定时任务

## rock.cron(string)
参数为定时任务名称
```lua
    local cron = rock.cron("metric")
    
    cron.task("@every 1s" , "获取系统信息1s" , function() end)
    cron.task("@every 5s" , "获取系统信息5s" , function() end)
    cron.task("@every 5s" , "获取系统信息5s" , function() end)
```

## cron.task(spec , title , function)
启动和添加任务函数
```lua
    cron.task("@every 1s" , "获取系统信息1s" , function() 
        --todo    
    end)
```