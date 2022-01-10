/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AlecAivazis/survey/v2"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type CzType struct {
	Type    string
	Message string
}

type CzCommit struct {
	Type    string
	Scope   string
	Subject string
	Body    string
	Footer  string
}

var typeOptions = []string{
	"feat:       新的功能",
	"fix:        修补错误",
	"docs:       文档修改",
	"style:      格式变化",
	"refactor:   重构代码",
	"perf:       性能提高",
	"test:       测试用例",
	"chore:      构建变动",
}

func GenerateCommit(czCommit *CzCommit) string {
	t := czCommit.Type
	t = strings.Split(t, ":")[0]
	commit := fmt.Sprintf(
		"%s(%s): %s\n",
		t,
		czCommit.Scope,
		czCommit.Subject,
	)
	if czCommit.Body != "" {
		commit += czCommit.Body
	}
	commit += "\n"
	if czCommit.Footer != "" {
		commit += czCommit.Footer
	}
	return commit
}

func GitCommit(commit string, amend bool) (err error) {
	tempFile, err := ioutil.TempFile("", "git_commit_")
	if err != nil {
		return
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()
	if _, err = tempFile.WriteString(commit); err != nil {
		return
	}
	args := []string{"commit"}
	if amend {
		args = append(args, "--amend")
	}
	args = append(args, "-F", tempFile.Name())
	cmd := exec.Command("git", args...)
	result, err := cmd.CombinedOutput()
	if err != nil && !strings.ContainsAny(err.Error(), "exit status") {
		return
	} else {
		fmt.Println(string(bytes.TrimSpace(result)))
	}
	return nil
}

func TypeTransform(an interface{}) (newAn interface{}) {
	if an != nil {
		if v, ok := an.(string); ok {
			return strings.Split(v, ":")[0]
		}
	}
	return an
}

// the questions to ask
var qs = []*survey.Question{
	{
		Name: "type",
		Prompt: &survey.Select{
			Message: "选择一个提交类型:",
			Options: typeOptions,
			Default: typeOptions[0],
		},
		Transform: TypeTransform,
		Validate:  survey.Required,
	},
	{
		Name:     "scope",
		Prompt:   &survey.Input{Message: "说明本次提交的影响范围:"},
		Validate: survey.Required,
	},
	{
		Name:     "subject",
		Prompt:   &survey.Input{Message: "对本次提交进行简短描述:"},
		Validate: survey.Required,
	},
}

var cfgFile string
var amend bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gcz",
	Short: "规范化 git commit message",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		// perform the questions
		czCommit := &CzCommit{}
		err := survey.Ask(qs, czCommit)
		if err != nil {
			log.Println(err.Error())
			return
		}
		commit := GenerateCommit(czCommit)
		if err := GitCommit(commit, amend); err != nil {
			log.Println(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gcz.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolP("amend", "a", false, "amend the last commit")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".gcz" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".gcz")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
