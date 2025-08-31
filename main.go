package main

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

func main() {
	// 修改注册表值
	err := modifyRegistry()
	if err != nil {
		fmt.Println("修改注册表失败:", err)
		waitForInput() // 等待用户输入
		return
	}

	// 提示信息
	fmt.Println("注册表已成功修改")
	waitForInput() // 等待用户输入
}

func modifyRegistry() error {
	// 打开注册表项
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\WindowsUpdate\UX\Settings`, registry.WRITE)
	if err != nil {
		return fmt.Errorf("无法打开注册表键: %w", err)
	}
	defer key.Close()

	// 设置注册表值
	err = key.SetDWordValue("FlightSettingsMaxPauseDays", 10000)
	if err != nil {
		return fmt.Errorf("设置DWORD值失败: %w", err)
	}
	return nil
}

// waitForInput waits for user input
func waitForInput() {
	fmt.Println("按 Enter 键以退出...")
	fmt.Scanln()
}
