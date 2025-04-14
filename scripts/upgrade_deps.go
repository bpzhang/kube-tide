package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// 命令行选项
var (
	autoUpgrade    bool   // 自动升级所有小版本更新
	autoApprove    bool   // 自动确认所有升级
	patchOnly      bool   // 仅升级补丁版本
	skipPrompt     bool   // 跳过用户交互提示
	skipTests      bool   // 跳过测试
	excludeModules string // 排除的模块，逗号分隔
)

func init() {
	flag.BoolVar(&autoUpgrade, "auto", false, "自动升级所有小版本更新")
	flag.BoolVar(&autoApprove, "y", false, "自动确认所有升级")
	flag.BoolVar(&patchOnly, "patch", false, "仅升级补丁版本")
	flag.BoolVar(&skipPrompt, "no-prompt", false, "跳过用户交互提示")
	flag.BoolVar(&skipTests, "no-tests", false, "跳过测试")
	flag.StringVar(&excludeModules, "exclude", "", "排除的模块列表，逗号分隔")

	// 添加使用说明
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "用法: %s [选项]\n\n选项:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  %s -auto -y        # 自动升级所有小版本，无需确认\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -patch          # 仅升级补丁版本\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -exclude=github.com/pkg/errors,golang.org/x/sys # 排除指定模块\n", os.Args[0])
	}
}

func main() {
	flag.Parse()

	// 备份当前的 go.mod 和 go.sum
	fmt.Println("备份当前依赖文件...")
	copyFile("go.mod", "go.mod.backup")
	copyFile("go.sum", "go.sum.backup")

	// 列出可以升级的依赖
	fmt.Println("检查可升级的依赖...")
	outdatedDeps := getOutdatedDependencies()

	// 如果只升级补丁版本，过滤出补丁版本更新
	if patchOnly {
		outdatedDeps = filterPatchVersions(outdatedDeps)
	}

	// 过滤排除的模块
	if excludeModules != "" {
		excludeList := strings.Split(excludeModules, ",")
		outdatedDeps = filterExcludedModules(outdatedDeps, excludeList)
	}

	if len(outdatedDeps) == 0 {
		fmt.Println("所有依赖已是最新版本！")
		return
	}

	fmt.Printf("发现 %d 个可升级的依赖:\n", len(outdatedDeps))
	for i, dep := range outdatedDeps {
		fmt.Printf("%d. %s: %s -> %s\n", i+1, dep.name, dep.currentVersion, dep.latestVersion)
	}

	// 一键自动升级
	if autoUpgrade {
		upgradeAllDeps(outdatedDeps)
		return
	}

	// 交互式升级
	interactiveUpgrade(outdatedDeps)
}

// 过滤仅保留补丁版本更新的依赖
func filterPatchVersions(deps []Dependency) []Dependency {
	var filtered []Dependency

	for _, dep := range deps {
		// 检查是否只有补丁版本变化
		if isPatchVersionChange(dep.currentVersion, dep.latestVersion) {
			filtered = append(filtered, dep)
		}
	}

	return filtered
}

// 判断是否是补丁版本变更
func isPatchVersionChange(current, latest string) bool {
	// 去除前缀
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// 分割版本号
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	// 确保有足够的部分来比较
	if len(currentParts) < 3 || len(latestParts) < 3 {
		return false
	}

	// 主版本和次版本必须相同，只有补丁版本不同
	return currentParts[0] == latestParts[0] &&
		currentParts[1] == latestParts[1] &&
		currentParts[2] != latestParts[2]
}

// 过滤掉被排除的模块
func filterExcludedModules(deps []Dependency, excludeList []string) []Dependency {
	var filtered []Dependency

	for _, dep := range deps {
		excluded := false
		for _, exclude := range excludeList {
			if strings.TrimSpace(exclude) == dep.name {
				excluded = true
				break
			}
		}

		if !excluded {
			filtered = append(filtered, dep)
		}
	}

	return filtered
}

// 一键自动升级所有依赖
func upgradeAllDeps(deps []Dependency) {
	fmt.Println("\n开始自动升级所有依赖...")
	upgradedPackages := []string{}

	for i, dep := range deps {
		fmt.Printf("\n[%d/%d] 升级 %s 从 %s 到 %s\n", i+1, len(deps), dep.name, dep.currentVersion, dep.latestVersion)

		// 执行升级
		if upgradePackage(dep.name, dep.latestVersion) {
			upgradedPackages = append(upgradedPackages, dep.name)
		}

		// 每3个包后运行测试(除非跳过测试)
		if !skipTests && ((i+1)%3 == 0 || i == len(deps)-1) {
			if runTests() {
				fmt.Println("测试通过！")
			} else {
				fmt.Println("测试失败！恢复备份...")
				restoreBackup()
				return
			}
		}
	}

	// 运行 go mod tidy
	fmt.Println("\n正在运行 go mod tidy...")
	execCommand("go", "mod", "tidy")

	// 检查依赖冲突
	fmt.Println("\n检查依赖冲突...")
	if !checkDependencyConflicts() {
		fmt.Println("检测到依赖冲突！恢复备份...")
		restoreBackup()
		return
	}

	// 最终测试
	if !skipTests {
		fmt.Println("\n运行最终测试...")
		if runTests() {
			fmt.Println("所有测试通过！升级成功完成。")
		} else {
			fmt.Println("最终测试失败！恢复备份...")
			restoreBackup()
			return
		}
	}

	// 显示升级结果
	if len(upgradedPackages) > 0 {
		fmt.Println("\n成功升级的包:")
		for i, pkg := range upgradedPackages {
			fmt.Printf("%d. %s\n", i+1, pkg)
		}
	}

	fmt.Println("你可以删除备份文件 go.mod.backup 和 go.sum.backup")
}

// 交互式升级依赖
func interactiveUpgrade(deps []Dependency) {
	scanner := bufio.NewScanner(os.Stdin)
	upgradedPackages := []string{}

	for i, dep := range deps {
		// 如果设置了自动确认，则跳过提示
		var answer string
		if autoApprove {
			answer = "y"
			fmt.Printf("\n自动确认升级 %s 从 %s 到 %s\n", dep.name, dep.currentVersion, dep.latestVersion)
		} else {
			fmt.Printf("\n是否升级 %s 从 %s 到 %s? (y/n/q): ", dep.name, dep.currentVersion, dep.latestVersion)
			scanner.Scan()
			answer = strings.ToLower(scanner.Text())
		}

		if answer == "q" {
			fmt.Println("退出升级过程。")
			return
		}

		if answer == "y" {
			if upgradePackage(dep.name, dep.latestVersion) {
				upgradedPackages = append(upgradedPackages, dep.name)
			}
		} else {
			fmt.Printf("跳过 %s 的升级\n", dep.name)
		}

		// 每3个包或最后一个包后运行测试(除非跳过测试)
		if !skipTests && ((i+1)%3 == 0 || i == len(deps)-1) {
			if runTests() {
				fmt.Println("测试通过！")
			} else {
				fmt.Println("测试失败！恢复备份...")
				restoreBackup()
				return
			}
		}
	}

	// 运行 go mod tidy
	fmt.Println("\n正在运行 go mod tidy...")
	execCommand("go", "mod", "tidy")

	// 检查依赖冲突
	fmt.Println("\n检查依赖冲突...")
	if !checkDependencyConflicts() {
		fmt.Println("检测到依赖冲突！恢复备份...")
		restoreBackup()
		return
	}

	// 最终测试
	if !skipTests {
		fmt.Println("\n运行最终测试...")
		if runTests() {
			fmt.Println("所有测试通过！升级成功完成。")
			if len(upgradedPackages) > 0 {
				fmt.Println("\n成功升级的包:")
				for i, pkg := range upgradedPackages {
					fmt.Printf("%d. %s\n", i+1, pkg)
				}
			}
			fmt.Println("你可以删除备份文件 go.mod.backup 和 go.sum.backup")
		} else {
			fmt.Println("最终测试失败！恢复备份...")
			restoreBackup()
		}
	} else {
		fmt.Println("\n跳过最终测试，升级完成。")
		if len(upgradedPackages) > 0 {
			fmt.Println("\n成功升级的包:")
			for i, pkg := range upgradedPackages {
				fmt.Printf("%d. %s\n", i+1, pkg)
			}
		}
		fmt.Println("你可以删除备份文件 go.mod.backup 和 go.sum.backup")
	}
}

type Dependency struct {
	name           string
	currentVersion string
	latestVersion  string
}

// 获取过时的依赖列表
func getOutdatedDependencies() []Dependency {
	// 不再使用 go get -u=patch，而是直接检查可用更新
	fmt.Println("检查模块更新...")

	// 保存原始依赖关系图，用于后续兼容性检查
	depGraph := getDependencyGraph()

	// 使用 go list -u -m all 检查所有可用更新
	cmd := exec.Command("go", "list", "-u", "-m", "all")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("检查依赖时出错: %v\n", err)
		os.Exit(1)
	}

	var outdated []Dependency
	scanner := bufio.NewScanner(&out)

	// 添加调试信息
	fmt.Println("分析依赖更新信息...")
	debugLines := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		debugLines = append(debugLines, line)

		// 打印原始输出以便调试
		// fmt.Println("DEBUG:", line)

		parts := strings.Fields(line)

		// 格式为: module v1.0.0 [v1.1.0]
		if len(parts) < 3 {
			continue
		}

		name := parts[0]
		currentVersion := parts[1]

		// 检查是否有可用更新
		var latestVersion string
		for i := 2; i < len(parts); i++ {
			if strings.HasPrefix(parts[i], "[") && strings.HasSuffix(parts[i], "]") {
				latestVersion = strings.Trim(parts[i], "[]")
				break
			}
		}

		// 如果没有找到最新版本，或者最新版本与当前版本相同，跳过
		if latestVersion == "" || latestVersion == currentVersion {
			continue
		}

		// 规范化版本格式
		if !strings.HasPrefix(latestVersion, "v") && strings.HasPrefix(currentVersion, "v") {
			latestVersion = "v" + latestVersion
		}

		// 避免主版本升级
		if !isMajorVersionChange(currentVersion, latestVersion) &&
			name != "golang.org/x/tools" &&
			!strings.Contains(name, "internal") &&
			!strings.HasPrefix(name, "github.com/spf13/viper") { // 避免某些特定包的升级

			// 确保版本确实是更新的
			if compareVersions(currentVersion, latestVersion) < 0 {
				outdated = append(outdated, Dependency{name, currentVersion, latestVersion})
			}
		}
	}

	// 如果没有找到可更新的依赖，输出调试信息
	if len(outdated) == 0 {
		fmt.Println("\n没有找到可升级的依赖。调试信息:")
		for _, line := range debugLines {
			if strings.Contains(line, "[") {
				fmt.Println("  - " + line)
			}
		}
		fmt.Println("\n如果你确定有可用更新，可以尝试以下操作:")
		fmt.Println("1. 运行 'go mod tidy' 清理依赖")
		fmt.Println("2. 确保你在项目根目录（包含go.mod文件的目录）执行此脚本")
	} else {
		// 按照依赖关系排序，优先升级底层依赖
		sortDependenciesByGraph(outdated, depGraph)
	}

	return outdated
}

// 获取当前项目的依赖关系图
func getDependencyGraph() map[string][]string {
	cmd := exec.Command("go", "mod", "graph")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("获取依赖关系图时出错: %v\n", err)
		return make(map[string][]string)
	}

	// 构建依赖关系图
	graph := make(map[string][]string)
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			// 格式: module1 module2 (module1 依赖 module2)
			depender := cleanModuleName(parts[0])
			dependency := cleanModuleName(parts[1])

			if _, exists := graph[depender]; !exists {
				graph[depender] = []string{}
			}
			graph[depender] = append(graph[depender], dependency)
		}
	}

	return graph
}

// 清理模块名称（移除版本信息）
func cleanModuleName(module string) string {
	if idx := strings.Index(module, "@"); idx != -1 {
		return module[:idx]
	}
	return module
}

// 根据依赖关系图排序依赖，底层依赖优先
func sortDependenciesByGraph(deps []Dependency, graph map[string][]string) {
	// 这里使用一个简单的启发式方法：
	// 被依赖越多的模块越底层，应该优先升级
	dependedCount := make(map[string]int)

	// 计算每个模块被依赖的次数
	for _, dependencies := range graph {
		for _, dep := range dependencies {
			dependedCount[dep]++
		}
	}

	// 按照被依赖次数排序
	stableSortDependencies(deps, func(i, j int) bool {
		countI := dependedCount[deps[i].name]
		countJ := dependedCount[deps[j].name]
		return countI > countJ // 被依赖多的排在前面
	})
}

// 稳定排序依赖
func stableSortDependencies(deps []Dependency, less func(i, j int) bool) {
	// 冒泡排序实现稳定排序（虽然效率不高，但对于小数据集足够）
	n := len(deps)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if less(j+1, j) {
				deps[j], deps[j+1] = deps[j+1], deps[j]
			}
		}
	}
}

// 判断是否是主版本变更
func isMajorVersionChange(current, latest string) bool {
	// 去除可能的前缀，比如 'v', 'go'
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// 分割版本号为组件部分
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	// 确保我们有至少一个部分来比较
	if len(currentParts) > 0 && len(latestParts) > 0 {
		// 主版本不同则是主版本变更
		return currentParts[0] != latestParts[0]
	}

	return false
}

// 比较版本号，返回:
// -1 如果 v1 < v2
//
//	0 如果 v1 == v2
//	1 如果 v1 > v2
func compareVersions(v1, v2 string) int {
	// 移除 'v' 前缀
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	// 比较每个版本部分
	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		var num1, num2 int
		fmt.Sscanf(v1Parts[i], "%d", &num1)
		fmt.Sscanf(v2Parts[i], "%d", &num2)

		if num1 < num2 {
			return -1
		} else if num1 > num2 {
			return 1
		}
	}

	// 如果前面的部分都相同，那么部分较多的版本更大
	if len(v1Parts) < len(v2Parts) {
		return -1
	} else if len(v1Parts) > len(v2Parts) {
		return 1
	}

	return 0
}

// 升级特定包，返回是否升级成功
func upgradePackage(name, version string) bool {
	// 保存当前状态以便回滚
	tempBackup()

	// 执行升级
	fmt.Printf("升级 %s 到 %s...\n", name, version)
	result := execCommandWithResult("go", "get", fmt.Sprintf("%s@%s", name, version))

	// 检查升级后的兼容性
	if !result.success {
		fmt.Printf("升级 %s 失败，正在回滚...\n", name)
		restoreTempBackup()
		return false
	}

	// 验证升级后的依赖兼容性
	if !skipTests {
		fmt.Println("验证依赖兼容性...")
		if !verifyDependencyCompatibility() {
			fmt.Printf("升级 %s 后依赖不兼容，正在回滚...\n", name)
			restoreTempBackup()
			return false
		}
	}

	// 升级成功，删除临时备份
	removeTempBackup()
	fmt.Printf("成功升级 %s 到 %s\n", name, version)
	return true
}

// 验证依赖兼容性
func verifyDependencyCompatibility() bool {
	// 运行 go mod tidy 检查依赖兼容性
	result := execCommandWithResult("go", "mod", "tidy", "-v")
	if !result.success {
		return false
	}

	// 运行 go mod verify 验证依赖校验和
	result = execCommandWithResult("go", "mod", "verify")
	if !result.success {
		fmt.Println("依赖校验和验证失败，可能存在版本冲突")
		return false
	}

	return true
}

// 创建临时备份用于单个包升级的回滚
func tempBackup() {
	copyFile("go.mod", "go.mod.temp")
	copyFile("go.sum", "go.sum.temp")
}

// 恢复临时备份
func restoreTempBackup() {
	copyFile("go.mod.temp", "go.mod")
	copyFile("go.sum.temp", "go.sum")
	removeTempBackup()
}

// 删除临时备份
func removeTempBackup() {
	os.Remove("go.mod.temp")
	os.Remove("go.sum.temp")
}

// 执行命令并返回结果
type execResult struct {
	success bool
	output  string
}

func execCommandWithResult(name string, args ...string) execResult {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("执行命令 %s %s 出错: %v\n", name, strings.Join(args, " "), err)
		return execResult{success: false}
	}

	return execResult{success: true, output: out.String()}
}

// 检查依赖冲突
func checkDependencyConflicts() bool {
	// 验证模块的完整性
	verifyResult := execCommandWithResult("go", "mod", "verify")
	if !verifyResult.success {
		fmt.Println("模块验证失败，可能存在依赖冲突:")
		fmt.Println(verifyResult.output)
		return false
	}

	// 检查是否有重复的依赖版本
	whyResult := execCommandWithResult("go", "mod", "why", "-m", "all")
	if !whyResult.success {
		fmt.Println("依赖分析失败:")
		fmt.Println(whyResult.output)
		return false
	}

	// 输出最终的依赖图供用户检查
	fmt.Println("当前依赖关系:")
	graphResult := execCommandWithResult("go", "mod", "graph")
	if graphResult.success {
		// 简化输出，只显示有问题的依赖
		lines := strings.Split(graphResult.output, "\n")
		duplicateVersions := findDuplicateVersions(lines)

		if len(duplicateVersions) > 0 {
			fmt.Println("警告: 检测到以下模块有多个版本:")
			for module, versions := range duplicateVersions {
				fmt.Printf("- %s: %s\n", module, strings.Join(versions, ", "))
			}

			// 如果是自动模式或设置了自动确认，自动继续
			if autoUpgrade || autoApprove || skipPrompt {
				fmt.Println("自动模式 - 忽略冲突并继续")
				return true
			}

			// 否则提示用户确认
			fmt.Print("是否继续? (y/n): ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			return strings.ToLower(scanner.Text()) == "y"
		}
	}

	return true
}

// 查找重复版本的依赖
func findDuplicateVersions(graphLines []string) map[string][]string {
	moduleVersions := make(map[string]map[string]bool)

	for _, line := range graphLines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		for _, part := range parts {
			if idx := strings.Index(part, "@"); idx != -1 {
				module := part[:idx]
				version := part[idx+1:]

				if _, exists := moduleVersions[module]; !exists {
					moduleVersions[module] = make(map[string]bool)
				}
				moduleVersions[module][version] = true
			}
		}
	}

	// 只返回有多个版本的模块
	duplicates := make(map[string][]string)
	for module, versions := range moduleVersions {
		if len(versions) > 1 {
			versionList := []string{}
			for version := range versions {
				versionList = append(versionList, version)
			}
			duplicates[module] = versionList
		}
	}

	return duplicates
}

// 运行测试，带有详细的失败报告
func runTests() bool {
	fmt.Println("运行测试并检查兼容性...")

	// 首先构建项目
	buildResult := execCommandWithResult("go", "build", "./...")
	if !buildResult.success {
		fmt.Println("构建失败，可能存在不兼容的API变更:")
		fmt.Println(buildResult.output)
		return false
	}

	// 然后运行测试
	testResult := execCommandWithResult("go", "test", "./...")
	if !testResult.success {
		fmt.Println("测试失败，可能存在兼容性问题:")
		fmt.Println(testResult.output)
		return false
	}

	return true
}

// 恢复备份
func restoreBackup() {
	copyFile("go.mod.backup", "go.mod")
	copyFile("go.sum.backup", "go.sum")
	fmt.Println("已恢复到原始依赖版本")
}

// 复制文件
func copyFile(src, dst string) {
	data, err := os.ReadFile(src)
	if err != nil {
		fmt.Printf("读取文件 %s 出错: %v\n", src, err)
		return
	}

	err = os.WriteFile(dst, data, 0644)
	if err != nil {
		fmt.Printf("写入文件 %s 出错: %v\n", dst, err)
		return
	}
}

// 执行命令并显示输出
func execCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("执行命令 %s %s 出错: %v\n", name, strings.Join(args, " "), err)
	}
}
