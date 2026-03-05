package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"voicepaper/config"
)

func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	// QQ邮箱SMTP配置 (从配置中获取)
	smtpHost := cfg.Auth.Email.SMTPHost
	smtpPort := cfg.Auth.Email.SMTPPort
	smtpUser := cfg.Auth.Email.SMTPUser
	smtpPassword := cfg.Auth.Email.SMTPPassword
	fromEmail := cfg.Auth.Email.FromEmail
	toEmail := smtpUser // Send to self for testing

	// 测试SMTP连接
	fmt.Println("🔍 测试QQ邮箱SMTP连接...")
	fmt.Printf("服务器: %s:%d\n", smtpHost, smtpPort)
	fmt.Printf("用户名: %s\n", smtpUser)
	fmt.Printf("授权码: %s (length: %d)\n", "*****", len(smtpPassword))
	fmt.Println()

	// 方法1：使用STARTTLS（端口587）
	fmt.Println("📧 方法1：使用STARTTLS（端口587）")
	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)

	// 建立连接
	client, err := smtp.Dial(addr)
	if err != nil {
		fmt.Printf("❌ 连接失败: %v\n", err)
		return
	}
	defer client.Close()
	fmt.Println("✅ 连接成功")

	// 发送STARTTLS命令
	if ok, _ := client.Extension("STARTTLS"); ok {
		fmt.Println("✅ 服务器支持STARTTLS")
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         smtpHost,
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			fmt.Printf("❌ STARTTLS失败: %v\n", err)
			return
		}
		fmt.Println("✅ STARTTLS成功")
	} else {
		fmt.Println("⚠️  服务器不支持STARTTLS")
	}

	// 认证
	fmt.Println("🔐 尝试认证...")
	if err = client.Auth(auth); err != nil {
		fmt.Printf("❌ 认证失败: %v\n", err)
		fmt.Println("\n可能的原因：")
		fmt.Println("1. 授权码不正确或已过期")
		fmt.Println("2. QQ邮箱SMTP服务未开启")
		fmt.Println("3. 账号异常或被限制")
		return
	}
	fmt.Println("✅ 认证成功")

	// 设置发件人
	if err = client.Mail(fromEmail); err != nil {
		fmt.Printf("❌ 设置发件人失败: %v\n", err)
		return
	}
	fmt.Println("✅ 设置发件人成功")

	// 设置收件人
	if err = client.Rcpt(toEmail); err != nil {
		fmt.Printf("❌ 设置收件人失败: %v\n", err)
		return
	}
	fmt.Println("✅ 设置收件人成功")

	// 发送邮件内容
	writer, err := client.Data()
	if err != nil {
		fmt.Printf("❌ 准备发送数据失败: %v\n", err)
		return
	}

	msg := []byte("Subject: 测试邮件\r\n\r\n这是一封测试邮件。")
	_, err = writer.Write(msg)
	if err != nil {
		fmt.Printf("❌ 写入邮件内容失败: %v\n", err)
		writer.Close()
		return
	}

	err = writer.Close()
	if err != nil {
		fmt.Printf("❌ 关闭数据流失败: %v\n", err)
		return
	}

	client.Quit()
	fmt.Println("✅ 邮件发送成功！")
}

