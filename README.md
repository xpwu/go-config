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
