package main

import (
	"fmt"
	"os/exec"
	"time"

	"golang.org/x/sys/windows/registry"
)

func main() {
	// 程序开始提示
	fmt.Println("=== Windows更新暂停设置工具 ===")
	fmt.Println("本工具将帮助您延长Windows更新暂停天数的最大限制")
	fmt.Println("\n注意：请确保您以管理员身份运行此程序")
	waitForInput() // 等待用户确认

	// 修改注册表值
	fmt.Println("\n步骤1/3: 正在修改系统注册表...")
	err := modifyRegistry()
	if err != nil {
		fmt.Println("修改注册表失败:", err)
		fmt.Println("可能原因:")
		fmt.Println("1. 未以管理员身份运行程序")
		fmt.Println("2. 系统组策略限制了注册表修改")
		fmt.Println("3. 注册表路径在您的系统中不存在")
		waitForInput()
		return
	}

	// 提示信息
	fmt.Println("步骤1/3: 注册表已成功修改")
	time.Sleep(1 * time.Second)

	// 尝试打开Windows更新高级选项
	fmt.Println("\n步骤2/3: 尝试打开Windows更新高级设置...")
	success := openWindowsUpdateAdvancedSettings()

	if !success {
		fmt.Println("\n所有自动打开尝试均失败，将为您提供最可靠的手动方法:")
		fmt.Println("请严格按照以下步骤操作:")
		fmt.Println("1. 按下键盘上的 Win + R 组合键 (Win键即Windows图标键)")
		fmt.Println("2. 在弹出的\"运行\"对话框中，复制粘贴以下内容:")
		fmt.Println("   ms-settings:windowsupdate-advancedoptions")
		fmt.Println("3. 点击\"确定\"按钮或按下Enter键")

		fmt.Println("\n如果上述方法仍无效，请尝试:")
		fmt.Println("1. 按下 Win + S 打开搜索框")
		fmt.Println("2. 输入\"gpedit.msc\"并打开本地组策略编辑器")
		fmt.Println("3. 导航到: 计算机配置 -> 管理模板 -> Windows组件 -> Windows更新")
		fmt.Println("4. 在右侧找到并双击\"配置自动更新\"")
		fmt.Println("5. 选择\"已启用\"，然后在下方设置您希望的更新方式")
		waitForInput()
	}

	// 操作引导
	fmt.Println("\n步骤3/3: 配置暂停更新:")
	fmt.Println("1. 在打开的页面中，选择\"更新和安全\"-\"高级选项\"-找到\"暂停更新\"或\"暂停直到\"选项")
	fmt.Println("2. 选择您希望暂停更新的截止日期")
	fmt.Println("3. 确认设置并关闭窗口")
	fmt.Println("\n完成后请返回此程序按Enter键退出")
	waitForInput()

	fmt.Println("\n=== 操作完成 ===")
	fmt.Println("程序将在3秒后退出...")
	time.Sleep(3 * time.Second)
}

func modifyRegistry() error {
	// 尝试打开或创建注册表项（增加兼容性）
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\WindowsUpdate\UX\Settings`,
		registry.WRITE)
	if err != nil {
		return fmt.Errorf("无法打开/创建注册表键: %w", err)
	}
	defer key.Close()

	// 设置注册表值
	err = key.SetDWordValue("FlightSettingsMaxPauseDays", 10000)
	if err != nil {
		return fmt.Errorf("设置注册表值失败: %w", err)
	}

	// 添加额外的注册表项以增强效果
	err = key.SetDWordValue("PauseFeatureUpdatesStartTime", 0)
	if err != nil {
		fmt.Println("警告: 未能设置额外注册表项，但主设置已完成")
	}

	return nil
}

// 安全执行命令的函数，确保进程正确回收
func safeExecCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	// 启动命令
	if err := cmd.Start(); err != nil {
		return err
	}

	// 设置超时，防止命令无响应
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		// 命令正常完成
		return err
	case <-time.After(5 * time.Second):
		// 超时后强制终止进程
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("命令超时并无法终止: %w", err)
		}
		return fmt.Errorf("命令执行超时")
	}
}

// 尝试多种系统级方法打开Windows更新高级选项
func openWindowsUpdateAdvancedSettings() bool {
	// 方法1: 基础CMD命令
	if err := safeExecCommand("cmd.exe", "/c", "start ms-settings:windowsupdate-advancedoptions"); err == nil {
		fmt.Println("使用CMD命令成功打开高级设置")
		return true
	}

	// 方法2: PowerShell命令（管理员模式）
	if err := safeExecCommand("powershell.exe", "-Command",
		"Start-Process cmd -ArgumentList '/c start ms-settings:windowsupdate-advancedoptions' -Verb RunAs"); err == nil {
		fmt.Println("使用管理员PowerShell成功打开高级设置")
		return true
	}

	// 方法3: 直接调用系统设置进程
	if err := safeExecCommand("explorer.exe", "ms-settings:windowsupdate-advancedoptions"); err == nil {
		fmt.Println("使用资源管理器成功打开高级设置")
		return true
	}

	// 方法4: 先打开设置再导航
	if err := safeExecCommand("cmd.exe", "/c", "start ms-settings:windowsupdate"); err == nil {
		fmt.Println("已打开Windows更新主页，请手动点击\"高级选项\"")
		return true
	}

	// 方法5: 使用命令提示符全屏打开
	if err := safeExecCommand("cmd.exe", "/c", "start /max ms-settings:windowsupdate-advancedoptions"); err == nil {
		return true
	}

	return false
}

func waitForInput() {
	fmt.Print("请按 Enter 键继续...")
	var input string
	fmt.Scanln(&input)
}
