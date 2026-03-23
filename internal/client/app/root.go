package app

import (
	"fmt"
	"os"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/client/api"
	"github.com/arvaliullin/goph-keeper/internal/client/config"
	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	buildVersion string
	buildDate    string

	serverURL   string
	loginStr    string
	passwordStr string
	secretType  string
	secretData  string
	secretMeta  string
	secretFile  string
	outputFile  string
)

// SetBuildInfo устанавливает информацию о сборке (вызывается из main).
func SetBuildInfo(version, date string) {
	buildVersion = version
	buildDate = date
}

func getAPIClient() (*api.Client, *config.ClientConfig, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	if serverURL != "" {
		cfg.Server = serverURL
	}

	client := api.NewClient(cfg.Server)
	if cfg.Token != "" {
		client.SetToken(cfg.Token)
	}

	return client, cfg, nil
}

func resetCommandState() {
	serverURL = ""
	loginStr = ""
	passwordStr = ""
	secretType = ""
	secretData = ""
	secretMeta = ""
	secretFile = ""
	outputFile = ""
}

func loadPassword() (string, error) {
	if passwordStr != "" {
		return passwordStr, nil
	}

	fmt.Fprint(os.Stderr, "Password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}

	return string(password), nil
}

func saveToken(cfg *config.ClientConfig, token string) error {
	cfg.Token = token
	return config.SaveConfig(cfg)
}

func syncClientState(cfg *config.ClientConfig) error {
	cfg.SetLastSync(time.Now())
	return config.SaveConfig(cfg)
}

func printSecret(secret *domain.Secret) {
	fmt.Printf("ID: %s\n", secret.ID)
	fmt.Printf("Type: %s\n", secret.Type)
	if secret.DeletedAt != nil {
		fmt.Printf("Deleted at: %s\n", secret.DeletedAt.Format(time.RFC3339))
		return
	}
	if secret.Type == domain.SecretTypeBinary {
		fmt.Println("Data: <binary payload>")
	} else {
		fmt.Printf("Data: %s\n", string(secret.Data))
	}
	fmt.Printf("Metadata: %s\n", string(secret.Metadata))
}

func saveBinaryOutput(path string, data []byte) error {
	return os.WriteFile(path, data, 0600)
}

// Execute запускает корневую команду CLI.
func Execute() {
	resetCommandState()

	var rootCmd = &cobra.Command{
		Use:   "keeper",
		Short: "GophKeeper - менеджер паролей",
		Long:  `GophKeeper позволяет безопасно хранить и управлять вашими приватными данными.`,
	}

	rootCmd.PersistentFlags().StringVarP(&serverURL, "server", "s", "", "Server URL (overrides config)")

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Вывести версию клиента",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Build version: %s\n", buildVersion)
			fmt.Printf("Build date: %s\n", buildDate)
		},
	}

	var registerCmd = &cobra.Command{
		Use:   "register",
		Short: "Зарегистрировать нового пользователя",
		Run: func(cmd *cobra.Command, args []string) {
			client, cfg, err := getAPIClient()
			if err != nil {
				fmt.Println(err)
				return
			}
			password, err := loadPassword()
			if err != nil {
				fmt.Printf("Ошибка чтения пароля: %v\n", err)
				return
			}
			token, err := client.Register(loginStr, password)
			if err != nil {
				fmt.Printf("Ошибка регистрации: %v\n", err)
				return
			}
			if err := saveToken(cfg, token); err != nil {
				fmt.Printf("Ошибка сохранения конфига: %v\n", err)
			}
			fmt.Println("Успешная регистрация!")
		},
	}
	registerCmd.Flags().StringVarP(&loginStr, "login", "l", "", "Логин")
	registerCmd.Flags().StringVarP(&passwordStr, "password", "p", "", "Пароль (если не указан, будет запрошен интерактивно)")
	errLogin := registerCmd.MarkFlagRequired("login")
	if errLogin != nil {
		fmt.Printf("Error marking flag required: %v\n", errLogin)
	}

	var loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Войти в систему",
		Run: func(cmd *cobra.Command, args []string) {
			client, cfg, err := getAPIClient()
			if err != nil {
				fmt.Println(err)
				return
			}
			password, err := loadPassword()
			if err != nil {
				fmt.Printf("Ошибка чтения пароля: %v\n", err)
				return
			}
			token, err := client.Login(loginStr, password)
			if err != nil {
				fmt.Printf("Ошибка входа: %v\n", err)
				return
			}
			if err := saveToken(cfg, token); err != nil {
				fmt.Printf("Ошибка сохранения конфига: %v\n", err)
			}
			fmt.Println("Успешный вход!")
		},
	}
	loginCmd.Flags().StringVarP(&loginStr, "login", "l", "", "Логин")
	loginCmd.Flags().StringVarP(&passwordStr, "password", "p", "", "Пароль (если не указан, будет запрошен интерактивно)")
	errLogin = loginCmd.MarkFlagRequired("login")
	if errLogin != nil {
		fmt.Printf("Error marking flag required: %v\n", errLogin)
	}

	var addCmd = &cobra.Command{
		Use:   "add",
		Short: "Добавить новый секрет",
		Run: func(cmd *cobra.Command, args []string) {
			client, _, err := getAPIClient()
			if err != nil {
				fmt.Println(err)
				return
			}
			if domain.SecretType(secretType) == domain.SecretTypeBinary {
				fmt.Println("Для бинарных данных используйте команду add-binary")
				return
			}
			secret, err := client.CreateSecret(domain.SecretType(secretType), []byte(secretData), []byte(secretMeta))
			if err != nil {
				fmt.Printf("Ошибка добавления секрета: %v\n", err)
				return
			}
			fmt.Printf("Секрет добавлен, ID: %s\n", secret.ID)
			if _, cfg, cfgErr := getAPIClient(); cfgErr == nil {
				if err := syncClientState(cfg); err != nil {
					fmt.Printf("Ошибка сохранения sync-состояния: %v\n", err)
				}
			}
		},
	}
	addCmd.Flags().StringVarP(&secretType, "type", "t", "login", "Тип секрета (login, text, binary, card)")
	addCmd.Flags().StringVarP(&secretData, "data", "d", "", "Данные секрета")
	addCmd.Flags().StringVarP(&secretMeta, "meta", "m", "", "Метаданные (JSON или текст)")
	errData := addCmd.MarkFlagRequired("data")
	if errData != nil {
		fmt.Printf("Error marking flag required: %v\n", errData)
	}

	var addBinaryCmd = &cobra.Command{
		Use:   "add-binary",
		Short: "Добавить бинарный секрет",
		Run: func(cmd *cobra.Command, args []string) {
			client, cfg, err := getAPIClient()
			if err != nil {
				fmt.Println(err)
				return
			}

			file, err := os.Open(secretFile)
			if err != nil {
				fmt.Printf("Ошибка открытия файла: %v\n", err)
				return
			}
			defer file.Close()

			info, err := file.Stat()
			if err != nil {
				fmt.Printf("Ошибка чтения файла: %v\n", err)
				return
			}

			secret, err := client.CreateSecret(domain.SecretTypeBinary, nil, []byte(secretMeta))
			if err != nil {
				fmt.Printf("Ошибка создания бинарного секрета: %v\n", err)
				return
			}

			if err := client.UploadBinary(secret.ID, file, info.Size()); err != nil {
				if delErr := client.DeleteSecret(secret.ID); delErr != nil {
					fmt.Printf("Предупреждение: не удалось откатить секрет %s: %v\n", secret.ID, delErr)
				}
				fmt.Printf("Ошибка загрузки бинарных данных: %v\n", err)
				return
			}

			if err := syncClientState(cfg); err != nil {
				fmt.Printf("Ошибка сохранения sync-состояния: %v\n", err)
			}
			fmt.Printf("Бинарный секрет добавлен, ID: %s\n", secret.ID)
		},
	}
	addBinaryCmd.Flags().StringVarP(&secretFile, "file", "f", "", "Путь к бинарному файлу")
	addBinaryCmd.Flags().StringVarP(&secretMeta, "meta", "m", "", "Метаданные бинарного секрета")
	errFile := addBinaryCmd.MarkFlagRequired("file")
	if errFile != nil {
		fmt.Printf("Error marking flag required: %v\n", errFile)
	}

	var updateCmd = &cobra.Command{
		Use:   "update [id]",
		Short: "Обновить существующий секрет",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, cfg, err := getAPIClient()
			if err != nil {
				fmt.Println(err)
				return
			}
			secret, err := client.UpdateSecret(args[0], []byte(secretData), []byte(secretMeta))
			if err != nil {
				fmt.Printf("Ошибка обновления секрета: %v\n", err)
				return
			}
			fmt.Printf("Секрет обновлен, ID: %s\n", secret.ID)
			if err := syncClientState(cfg); err != nil {
				fmt.Printf("Ошибка сохранения sync-состояния: %v\n", err)
			}
		},
	}
	updateCmd.Flags().StringVarP(&secretData, "data", "d", "", "Новые данные секрета")
	updateCmd.Flags().StringVarP(&secretMeta, "meta", "m", "", "Новые метаданные (JSON или текст)")

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "Список всех секретов",
		Run: func(cmd *cobra.Command, args []string) {
			client, _, err := getAPIClient()
			if err != nil {
				fmt.Println(err)
				return
			}
			secrets, err := client.ListSecrets()
			if err != nil {
				fmt.Printf("Ошибка получения списка: %v\n", err)
				return
			}
			fmt.Printf("Найдено секретов: %d\n", len(secrets))
			for _, s := range secrets {
				meta := string(s.Metadata)
				if len(meta) > 50 {
					meta = meta[:50] + "..."
				}
				if meta != "" {
					fmt.Printf("ID: %s | Type: %s | Meta: %s | Created: %s\n", s.ID, s.Type, meta, s.CreatedAt.Format("2006-01-02 15:04:05"))
				} else {
					fmt.Printf("ID: %s | Type: %s | Created: %s\n", s.ID, s.Type, s.CreatedAt.Format("2006-01-02 15:04:05"))
				}
			}
		},
	}

	var getCmd = &cobra.Command{
		Use:   "get [id]",
		Short: "Получить секрет по ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, _, err := getAPIClient()
			if err != nil {
				fmt.Println(err)
				return
			}

			secretID := args[0]
			secret, err := client.GetSecret(secretID)
			if err != nil {
				fmt.Printf("Ошибка получения секрета: %v\n", err)
				return
			}

			printSecret(secret)
			if secret.Type == domain.SecretTypeBinary && outputFile != "" {
				data, err := client.DownloadBinary(secretID)
				if err != nil {
					fmt.Printf("Ошибка скачивания бинарных данных: %v\n", err)
					return
				}
				if err := saveBinaryOutput(outputFile, data); err != nil {
					fmt.Printf("Ошибка сохранения файла: %v\n", err)
					return
				}
				fmt.Printf("Бинарные данные сохранены в %s\n", outputFile)
			}
		},
	}
	getCmd.Flags().StringVarP(&outputFile, "out", "o", "", "Путь для сохранения бинарных данных")

	var downloadBinaryCmd = &cobra.Command{
		Use:   "download-binary [id]",
		Short: "Скачать бинарный секрет в файл",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, _, err := getAPIClient()
			if err != nil {
				fmt.Println(err)
				return
			}
			data, err := client.DownloadBinary(args[0])
			if err != nil {
				fmt.Printf("Ошибка скачивания бинарных данных: %v\n", err)
				return
			}
			if err := saveBinaryOutput(outputFile, data); err != nil {
				fmt.Printf("Ошибка сохранения файла: %v\n", err)
				return
			}
			fmt.Printf("Файл сохранен в %s\n", outputFile)
		},
	}
	downloadBinaryCmd.Flags().StringVarP(&outputFile, "out", "o", "", "Путь для сохранения файла")
	errOutput := downloadBinaryCmd.MarkFlagRequired("out")
	if errOutput != nil {
		fmt.Printf("Error marking flag required: %v\n", errOutput)
	}

	var deleteCmd = &cobra.Command{
		Use:   "delete [id]",
		Short: "Удалить секрет по ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, cfg, err := getAPIClient()
			if err != nil {
				fmt.Println(err)
				return
			}
			if err := client.DeleteSecret(args[0]); err != nil {
				fmt.Printf("Ошибка удаления секрета: %v\n", err)
				return
			}
			if err := syncClientState(cfg); err != nil {
				fmt.Printf("Ошибка сохранения sync-состояния: %v\n", err)
			}
			fmt.Println("Секрет удален")
		},
	}

	var logoutCmd = &cobra.Command{
		Use:   "logout",
		Short: "Выйти из системы (удалить сохраненный токен)",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Printf("Ошибка загрузки конфига: %v\n", err)
				return
			}
			if serverURL != "" {
				cfg.Server = serverURL
			}
			cfg.Token = ""
			if err := config.SaveConfig(cfg); err != nil {
				fmt.Printf("Ошибка сохранения конфига: %v\n", err)
				return
			}
			fmt.Println("Выход выполнен")
		},
	}

	var syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Получить изменения с сервера с момента последней синхронизации",
		Run: func(cmd *cobra.Command, args []string) {
			client, cfg, err := getAPIClient()
			if err != nil {
				fmt.Println(err)
				return
			}

			lastSync, err := cfg.LastSyncTime()
			if err != nil {
				fmt.Printf("Ошибка чтения времени синхронизации: %v\n", err)
				return
			}

			var secrets []*domain.Secret
			if lastSync.IsZero() {
				secrets, err = client.ListSecrets()
			} else {
				secrets, err = client.SyncSecrets(lastSync)
			}
			if err != nil {
				fmt.Printf("Ошибка синхронизации: %v\n", err)
				return
			}

			for _, secret := range secrets {
				printSecret(secret)
				fmt.Println()
			}

			if err := syncClientState(cfg); err != nil {
				fmt.Printf("Ошибка сохранения sync-состояния: %v\n", err)
			}
			fmt.Printf("Получено изменений: %d\n", len(secrets))
		},
	}

	rootCmd.AddCommand(
		versionCmd,
		registerCmd,
		loginCmd,
		logoutCmd,
		addCmd,
		addBinaryCmd,
		updateCmd,
		listCmd,
		getCmd,
		downloadBinaryCmd,
		deleteCmd,
		syncCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
