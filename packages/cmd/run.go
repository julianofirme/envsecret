package cmd

import (
	"envsecret/packages/api"
	"envsecret/packages/util"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [flags] -- [command]",
	Short: "Inject environment variables into your application process",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("at least one argument is required after the run command")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig(".envsecret.json")
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		userCreds, err := util.GetCurrentLoggedInUserDetails()
		if err != nil {
			util.HandleError(err, "Unable to get your login details")
		}

		if userCreds.LoginExpired {
			util.PrintErrorMessageAndExit("Your login session has expired, please run [envsecret login] and try again")
		}

		httpClient := resty.New()
		httpClient.SetAuthToken(userCreds.UserCredentials.JWTToken)

		secrets, err := api.CallGetSecrets(httpClient, config.WorkspaceId)
		if err != nil {
			fmt.Println("Error parsing secrets:", err)
			return
		}

		envVars := make(map[string]string)

		for _, s := range os.Environ() {
			kv := strings.SplitN(s, "=", 2)
			key := kv[0]
			value := kv[1]
			envVars[key] = value
		}

		for _, secret := range secrets.Secret {
			envVars[secret.Key] = secret.Value
		}

		var env []string
		for key, value := range envVars {
			s := key + "=" + value
			env = append(env, s)
		}

		command := args[0]
		argsForCommand := args[1:]

		cmdToRun := exec.Command(command, argsForCommand...)
		cmdToRun.Stdin = os.Stdin
		cmdToRun.Stdout = os.Stdout
		cmdToRun.Stderr = os.Stderr
		cmdToRun.Env = env

		sigChannel := make(chan os.Signal, 1)
		signal.Notify(sigChannel)

		if err := cmdToRun.Start(); err != nil {
			fmt.Println("Error starting command:", err)
			return
		}

		go func() {
			for {
				sig := <-sigChannel
				_ = cmdToRun.Process.Signal(sig)
			}
		}()

		if err := cmdToRun.Wait(); err != nil {
			_ = cmdToRun.Process.Signal(os.Kill)
			fmt.Println("Failed to wait for command termination:", err)
			return
		}

		waitStatus := cmdToRun.ProcessState.Sys().(syscall.WaitStatus)
		os.Exit(waitStatus.ExitStatus())
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
