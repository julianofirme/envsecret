package cmd

import (
	"encoding/base64"
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
		encryptionKey := string(os.Getenv("SECRET_ENCRYPTION_KEY"))
		if len(encryptionKey) == 0 {
			fmt.Println("SECRET_ENCRYPTION_KEY environment variable is required")
			return
		}

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
			fmt.Println("Error loading secrets file:", err)
			return
		}

		envVars := make(map[string]string)
		for _, secret := range secrets.Secret {
			// decrypt env key
			key_iv, err := base64.StdEncoding.DecodeString(secret.KeyIV)
			if err != nil {
				fmt.Println("unable to decode secret IV for secret key", err)
			}

			key_tag, err := base64.StdEncoding.DecodeString(secret.KeyAuthTag)
			if err != nil {
				fmt.Println("unable to decode secret authentication tag for secret key", err)
			}

			key_ciphertext, err := base64.StdEncoding.DecodeString(secret.KeyEncrypted)
			if err != nil {
				fmt.Println("unable to decode secret cipher text for secret key", err)
			}

			decryptedKey, err := util.DecryptSymmetric([]byte(encryptionKey), key_ciphertext, key_tag, key_iv)
			if err != nil {
				fmt.Println("Error decrypting secret key:", err)
				return
			}
			decryptedKeyString := string(decryptedKey)

			// decrypt env value
			value_iv, err := base64.StdEncoding.DecodeString(secret.ValueIV)
			if err != nil {
				fmt.Println("unable to decode secret IV for secret key", err)
			}

			value_tag, err := base64.StdEncoding.DecodeString(secret.ValueAuthTag)
			if err != nil {
				fmt.Println("unable to decode secret authentication tag for secret key", err)
			}

			value_ciphertext, err := base64.StdEncoding.DecodeString(secret.ValueEncrypted)
			if err != nil {
				fmt.Println("unable to decode secret cipher text for secret key", err)
			}

			decryptedValue, err := util.DecryptSymmetric([]byte(encryptionKey), value_ciphertext, value_tag, value_iv)
			if err != nil {
				fmt.Println("Error decrypting secret value:", err)
				return
			}
			decryptedValueString := string(decryptedValue)

			envVars[decryptedKeyString] = decryptedValueString
		}

		// merge with existing environment variables
		for _, s := range os.Environ() {
			kv := strings.SplitN(s, "=", 2)
			key := kv[0]
			value := kv[1]
			envVars[key] = value
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
