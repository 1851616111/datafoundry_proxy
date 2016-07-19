package main

var bccEmail []string = []string{
	"chaizs@asiainfo.com",
	"jingxy3@asiainfo.com",
	"xueys@maitewang.com",
	"jiangtong@asiainfo.com",
}

var Subject string = "Welcome to Datafoundry"
var Message1 string = `Hello %s, <br />please click <a href="%s">link</a> to verify your account, the activation link will be expire after 24 hours.`

var Message string = `<body style="background-color: #e6e9f2">
    <div class="sendemailbox">
        <div class="sendemail" style="margin:auto; background-color: white; width: 540px;">
            <div style="padding: 30px 30px">
                <div class="headLogo">
                    <img src="http://o9h84ok6f.bkt.clouddn.com/mail_banner.png_1468918176875.png" >
                </div>
                <div class="content" style="">
                    <p style="font-size: 16px; margin: 50px 0 15px 0px;color: #5a6378;">亲爱的 %s，感谢您注册<a style="font-size: 16px;color: #000000;">&nbsp;铸数工坊&nbsp;DataFoundry</a>。</p>
                    <p style="font-size: 16px;color: #5a6378; margin-bottom: 20px">请点击按钮激活您的账号。</p>
                    <div class="button" style="height: 100px;">
                        <a href="%s" ID="activation" type="button" style="height: 40px; width: 160px; font-size: 18px; background-color: #f6a540; border: 1px solid #f6a540; border-radius: 2px; margin-bottom: 60px; color: white; padding: 10px 20px; text-decoration:none">立即验证邮箱</a>
                    </div>
                    <div class="submes" style="font-size: 14px; color: #5a6378;">
                        <p>如果按钮无法点击，请将下面的链接复制到浏览器地址栏中打开：</p>
                        <p class="address" style="margin-top: 15px">%s</p>
                        <p style="margin-bottom: 30px">请您在 24 小时内激活。</p>
                    </div>
                </div>
            </div>

            <div class="buttom" style=" width:540px;">
                <div style="padding: 30px 30px; background-color: #f7f8fb;">
                    <img src="http://o9h84ok6f.bkt.clouddn.com/mail_qrcode.png_1468918176894.png">
                    <div style="padding-top: 0px; width:50%%; color: #000000;display: inline-block; margin-left: 20px; font-size: 12px">
                        <p style="; ">扫一扫</p>
                        <p style="margin-top: 10px">了解最新产品和咨询</p>
                        <p style="margin-top: 20px; color: #ef9033">铸数工坊公众号</p>
                    </div>
                </div>
            </div>
        </div>
        <div style="margin:0 auto; padding-top: 30px; padding-bottom: 35px; width: 540px; position:relative">
            <div class="line">
            </div>
            <img id="footer" src="http://o9h84ok6f.bkt.clouddn.com/mail_logo_small.png_1468918176890.png">
        </div>
    </div>

</body>
<style>
    #activation :hover {
        background-color: #e5993a;
    }
    #activation :active {
        background-color: #f8b551;
    }
    .line {
        width:100%%;
        border-bottom: 1px solid #c9d0e2;
    }
    #footer {
        background-color: #e6e9f2;
        position: absolute;
        top: 15px;
        left:45%%;
        height: 30px;
        width:30px;
        z-index: 5;
        padding: 0px 10px;
    }
</style>`
