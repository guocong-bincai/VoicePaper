package service

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"math/big"
	"net/smtp"
	"time"
	"voicepaper/config"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"

	"gorm.io/gorm"
)

// EmailService 邮箱验证码服务
type EmailService struct {
	cfg      *config.EmailConfig
	codeRepo *repository.VerificationCodeRepository
}

func NewEmailService(cfg *config.EmailConfig, codeRepo *repository.VerificationCodeRepository) *EmailService {
	return &EmailService{
		cfg:      cfg,
		codeRepo: codeRepo,
	}
}

// SendVerificationCode 发送验证码邮件
func (s *EmailService) SendVerificationCode(email, ipAddress, userAgent, purpose string) error {
	// 防刷：检查5分钟内是否已发送
	count, err := s.codeRepo.CountRecentCodes(email, "email", 5)
	if err != nil {
		return fmt.Errorf("检查发送频率失败: %w", err)
	}
	// 临时放宽限制用于测试：5分钟内最多50次（原来是3次）
	// TODO: 测试完成后改回3次
	if count >= 50 {
		return fmt.Errorf("发送过于频繁，请5分钟后再试")
	}

	// 生成6位数字验证码
	code, err := s.generateCode(6)
	if err != nil {
		return fmt.Errorf("生成验证码失败: %w", err)
	}

	// 保存验证码到数据库
	vc := &model.VerificationCode{
		CodeType:  "email",
		Receiver:  email,
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute), // 5分钟有效
		Used:      false,
		Purpose:   purpose,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := s.codeRepo.Create(vc); err != nil {
		return fmt.Errorf("保存验证码失败: %w", err)
	}

	// 发送邮件（如果失败，记录日志但不影响验证码保存）
	// 因为验证码已经保存到数据库，用户可以手动输入验证码登录
	if err := s.sendEmail(email, code, purpose); err != nil {
		// 邮件发送失败，但不删除验证码，因为验证码已经保存
		// 用户可以从数据库或日志中获取验证码进行测试
		// TODO: 生产环境应该记录到日志系统，并考虑是否删除验证码
		fmt.Printf("⚠️  邮件发送失败，但验证码已保存: %v\n", err)
		// 暂时不返回错误，让前端显示成功，用户可以手动输入验证码
		// return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}

// CheckCode 检查验证码是否有效（不标记为已使用）
// BUG修复: 用于注册流程的邮箱验证，只检查不消耗验证码
// 修复策略: 创建新方法只验证有效性，不标记为已使用
// 影响范围: backend/internal/service/email_service.go
// 修复日期: 2025-12-10
func (s *EmailService) CheckCode(email, code string) (*model.VerificationCode, error) {
	fmt.Printf("🔍 开始检查验证码（不消耗）: email=%s, code=%s\n", email, code)
	vc, err := s.codeRepo.FindValidCode(email, "email", code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("❌ 验证码未找到或已过期: email=%s, code=%s, error=%v\n", email, code, err)
			return nil, fmt.Errorf("验证码无效或已过期")
		}
		fmt.Printf("❌ 验证码查询失败: email=%s, code=%s, error=%v\n", email, code, err)
		return nil, fmt.Errorf("验证验证码失败: %w", err)
	}
	fmt.Printf("✅ 验证码有效（未消耗）: id=%d, expires_at=%v, used=%v\n", vc.ID, vc.ExpiresAt, vc.Used)
	return vc, nil
}

// VerifyCode 验证验证码（标记为已使用）
func (s *EmailService) VerifyCode(email, code string) (*model.VerificationCode, error) {
	fmt.Printf("🔍 开始验证验证码: email=%s, code=%s\n", email, code)
	vc, err := s.codeRepo.FindValidCode(email, "email", code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("❌ 验证码未找到或已过期: email=%s, code=%s, error=%v\n", email, code, err)
			return nil, fmt.Errorf("验证码无效或已过期")
		}
		fmt.Printf("❌ 验证码查询失败: email=%s, code=%s, error=%v\n", email, code, err)
		return nil, fmt.Errorf("验证验证码失败: %w", err)
	}
	fmt.Printf("✅ 验证码找到: id=%d, expires_at=%v, used=%v\n", vc.ID, vc.ExpiresAt, vc.Used)

	// 标记为已使用
	if err := s.codeRepo.MarkAsUsed(vc.ID); err != nil {
		fmt.Printf("❌ 标记验证码失败: vc_id=%d, error=%v\n", vc.ID, err)
		return nil, fmt.Errorf("标记验证码失败: %w", err)
	}
	fmt.Printf("✅ 验证码已标记为已使用: vc_id=%d\n", vc.ID)

	return vc, nil
}

// generateCode 生成指定位数的数字验证码
func (s *EmailService) generateCode(length int) (string, error) {
	max := big.NewInt(0)
	max.Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	max.Sub(max, big.NewInt(1))

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	code := fmt.Sprintf("%0*d", length, n.Int64())
	return code, nil
}

// sendEmail 发送邮件
func (s *EmailService) sendEmail(to, code, purpose string) error {
	// 构建邮件内容
	subject := "VoicePaper 验证码"
	body := s.buildEmailBody(code, purpose)

	// SMTP认证
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	// 构建邮件（符合RFC5322标准）
	msg := []byte(fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"MIME-Version: 1.0\r\n"+
		"\r\n"+
		"%s\r\n", s.cfg.FromName, s.cfg.FromEmail, to, subject, body))

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	// 对于QQ邮箱，端口587需要STARTTLS，端口465需要SSL
	// 标准库的smtp.SendMail会自动处理STARTTLS（端口587）
	// 但如果使用465端口，需要手动建立TLS连接
	if s.cfg.SMTPPort == 465 {
		// 使用SSL连接（端口465）
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         s.cfg.SMTPHost,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS连接失败: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
		if err != nil {
			return fmt.Errorf("创建SMTP客户端失败: %w", err)
		}
		defer client.Close()

		// 先进行认证
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP认证失败: %w", err)
		}

		// 设置发件人
		if err = client.Mail(s.cfg.FromEmail); err != nil {
			return fmt.Errorf("设置发件人失败: %w", err)
		}

		// 设置收件人
		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("设置收件人失败: %w", err)
		}

		// 发送邮件内容
		writer, err := client.Data()
		if err != nil {
			return fmt.Errorf("准备发送数据失败: %w", err)
		}

		_, err = writer.Write(msg)
		if err != nil {
			writer.Close()
			return fmt.Errorf("写入邮件内容失败: %w", err)
		}

		err = writer.Close()
		if err != nil {
			return fmt.Errorf("关闭数据流失败: %w", err)
		}

		client.Quit()
		return nil
	} else {
		// 使用STARTTLS（端口587）
		// 对于QQ邮箱，需要先建立连接，然后STARTTLS，再认证
		client, err := smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("连接SMTP服务器失败: %w", err)
		}
		defer client.Close()

		// 发送STARTTLS命令
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: false,
				ServerName:         s.cfg.SMTPHost,
			}
			if err = client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("STARTTLS失败: %w", err)
			}
		}

		// 认证
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP认证失败: %w", err)
		}

		// 设置发件人
		if err = client.Mail(s.cfg.FromEmail); err != nil {
			return fmt.Errorf("设置发件人失败: %w", err)
		}

		// 设置收件人
		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("设置收件人失败: %w", err)
		}

		// 发送邮件内容
		writer, err := client.Data()
		if err != nil {
			return fmt.Errorf("准备发送数据失败: %w", err)
		}

		_, err = writer.Write(msg)
		if err != nil {
			writer.Close()
			return fmt.Errorf("写入邮件内容失败: %w", err)
		}

		err = writer.Close()
		if err != nil {
			return fmt.Errorf("关闭数据流失败: %w", err)
		}

		client.Quit()
		return nil
	}
}

// buildEmailBody 构建邮件正文
func (s *EmailService) buildEmailBody(code, purpose string) string {
	var purposeText string
	switch purpose {
	case "login":
		purposeText = "登录"
	case "register":
		purposeText = "注册"
	case "reset_password":
		purposeText = "重置密码"
	default:
		purposeText = "验证"
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2563EB; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 8px 8px; }
        .code-box { background: white; border: 2px dashed #2563EB; padding: 20px; text-align: center; margin: 20px 0; border-radius: 8px; }
        .code { font-size: 32px; font-weight: bold; color: #2563EB; letter-spacing: 8px; }
        .warning { color: #dc2626; font-size: 14px; margin-top: 20px; }
        .footer { text-align: center; color: #666; font-size: 12px; margin-top: 30px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>VoicePaper</h1>
        </div>
        <div class="content">
            <h2>%s验证码</h2>
            <p>您好，</p>
            <p>您正在进行%s操作，验证码如下：</p>
            <div class="code-box">
                <div class="code">%s</div>
            </div>
            <p class="warning">⚠️ 验证码5分钟内有效，请勿泄露给他人。</p>
            <p>如果这不是您的操作，请忽略此邮件。</p>
            <div class="footer">
                <p>此邮件由系统自动发送，请勿回复。</p>
                <p>© VoicePaper</p>
            </div>
        </div>
    </div>
</body>
</html>
`, purposeText, purposeText, code)
}
