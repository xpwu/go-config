# go-config

1、收集所有package的配置项，统一输出为一个配置文件  
2、从配置文件读取所有package需要的配置数据

## Usage  

#### main()中的流程
1、定义一个满足Configurator接口的对象，此包中已经实现了json配置格式的Configurator对象；  
2、SetConfigurator()设置此对象；  
3、调用Read()读取配置 或者Print()打印默认的配置文件；  
4、调用其他模块功能，在其他模块功能中即可使用配置值；  

#### 各package中的使用
1、在package中用struct定义需要的配置数据，并初始化，初始化的的值就是Print()的默认值；  
2、在package的init()方法中调用 Unmarshal()方法。  

## Note 
配置数据的赋值是在Read()中执行，而Read()都是在init()后执行，所以在init()获取不到真实的配置值。

## Tag

```
// tag: `conf:name,tips` => key = name; tips = tips
// tag: `conf:name,tips,tip2` => key = name; tips = tips,tip2
// tag: `conf:name` => key = name; tips = ""
// tag: `conf:name,` => key = name; tips = ""
// tag: `conf:,tips` => key = ""; tips = tips
// tag: `conf:,tips,tip2` => key = ""; tips = tips,tip2
// tag: `conf:,` => key = ""; tips = ""
// tag: `conf:` => key = ""; tips = ""

// tag: `conf:-` ignore
```


## jsonConfig  
默认实现了json格式的配置输出，${key}-tips 的值即为对应key的帮助信息

## Configurator
可以通过Configurator接口实现其他的配置输出格式；在main()中调用SetConfigurator()
设置具体的输出类
