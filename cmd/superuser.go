package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/spf13/cobra"
)

// NewSuperuserCommand creates and returns new command for managing
// superuser accounts (create, update, upsert, delete).
func NewSuperuserCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:   "superuser",
		Short: "Manage superusers",
	}

	command.AddCommand(superuserUpsertCommand(app))
	command.AddCommand(superuserCreateCommand(app))
	command.AddCommand(superuserUpdateCommand(app))
	command.AddCommand(superuserDeleteCommand(app))
	command.AddCommand(superuserOTPCommand(app))

	return command
}

func superuserUpsertCommand(app core.App) *cobra.Command {
	var role string

	command := &cobra.Command{
		Use:          "upsert",
		Example:      "superuser upsert test@example.com 1234567890\n  superuser upsert test@example.com 1234567890 --role admin",
		Short:        "Creates, or updates if email exists, a single superuser",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("missing email and password arguments")
			}

			if args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
				return errors.New("missing or invalid email address")
			}

			superusersCol, err := app.FindCachedCollectionByNameOrId(core.CollectionNameSuperusers)
			if err != nil {
				return fmt.Errorf("failed to fetch %q collection: %w", core.CollectionNameSuperusers, err)
			}

			role = normalizeSuperuserRoleFlag(role)
			if role != "" {
				if err := validateSuperuserRoleFlag(superusersCol, role); err != nil {
					return err
				}
			}

			superuser, err := app.FindAuthRecordByEmail(superusersCol, args[0])
			if err != nil {
				superuser = core.NewRecord(superusersCol)
			}

			superuser.SetEmail(args[0])
			superuser.SetPassword(args[1])
			if role != "" {
				superuser.Set("role", role)
			}

			if err := app.Save(superuser); err != nil {
				return fmt.Errorf("failed to upsert superuser account: %w", err)
			}

			color.Green("Successfully saved superuser %q!", superuser.Email())
			return nil
		},
	}

	command.Flags().StringVar(&role, "role", "", "optional Nuvio superuser role to assign (admin or client)")

	return command
}

func normalizeSuperuserRoleFlag(role string) string {
	return strings.TrimSpace(strings.ToLower(role))
}

func validateSuperuserRoleFlag(superusersCol *core.Collection, role string) error {
	if role != "admin" && role != "client" {
		return fmt.Errorf("invalid superuser role %q; expected \"admin\" or \"client\"", role)
	}

	if superusersCol == nil || superusersCol.Fields.GetByName("role") == nil {
		return fmt.Errorf("superuser role field is missing; run the Nuvio migrations before using --role")
	}

	selectField, ok := superusersCol.Fields.GetByName("role").(*core.SelectField)
	if !ok || len(selectField.Values) == 0 {
		return nil
	}

	for _, value := range selectField.Values {
		if strings.TrimSpace(strings.ToLower(value)) == role {
			return nil
		}
	}

	return fmt.Errorf("superuser role %q is not allowed by the %q schema", role, core.CollectionNameSuperusers)
}

func superuserCreateCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:          "create",
		Example:      "superuser create test@example.com 1234567890",
		Short:        "Creates a new superuser",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("missing email and password arguments")
			}

			if args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
				return errors.New("missing or invalid email address")
			}

			superusersCol, err := app.FindCachedCollectionByNameOrId(core.CollectionNameSuperusers)
			if err != nil {
				return fmt.Errorf("failed to fetch %q collection: %w", core.CollectionNameSuperusers, err)
			}

			superuser := core.NewRecord(superusersCol)
			superuser.SetEmail(args[0])
			superuser.SetPassword(args[1])

			if err := app.Save(superuser); err != nil {
				return fmt.Errorf("failed to create new superuser account: %w", err)
			}

			color.Green("Successfully created new superuser %q!", superuser.Email())
			return nil
		},
	}

	return command
}

func superuserUpdateCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:          "update",
		Example:      "superuser update test@example.com 1234567890",
		Short:        "Changes the password of a single superuser",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("missing email and password arguments")
			}

			if args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
				return errors.New("missing or invalid email address")
			}

			superuser, err := app.FindAuthRecordByEmail(core.CollectionNameSuperusers, args[0])
			if err != nil {
				return fmt.Errorf("superuser with email %q doesn't exist", args[0])
			}

			superuser.SetPassword(args[1])

			if err := app.Save(superuser); err != nil {
				return fmt.Errorf("failed to change superuser %q password: %w", superuser.Email(), err)
			}

			color.Green("Successfully changed superuser %q password!", superuser.Email())
			return nil
		},
	}

	return command
}

func superuserDeleteCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:          "delete",
		Example:      "superuser delete test@example.com",
		Short:        "Deletes an existing superuser",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 || args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
				return errors.New("invalid or missing email address")
			}

			superuser, err := app.FindAuthRecordByEmail(core.CollectionNameSuperusers, args[0])
			if err != nil {
				color.Yellow("superuser %q is missing or already deleted", args[0])
				return nil
			}

			if err := app.Delete(superuser); err != nil {
				return fmt.Errorf("failed to delete superuser %q: %w", superuser.Email(), err)
			}

			color.Green("Successfully deleted superuser %q!", superuser.Email())
			return nil
		},
	}

	return command
}

func superuserOTPCommand(app core.App) *cobra.Command {
	command := &cobra.Command{
		Use:          "otp",
		Example:      "superuser otp test@example.com",
		Short:        "Creates a new OTP for the specified superuser",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 || args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
				return errors.New("invalid or missing email address")
			}

			superuser, err := app.FindAuthRecordByEmail(core.CollectionNameSuperusers, args[0])
			if err != nil {
				return fmt.Errorf("superuser with email %q doesn't exist", args[0])
			}

			if !superuser.Collection().OTP.Enabled {
				return errors.New("OTP auth is not enabled for the _superusers collection")
			}

			pass := security.RandomStringWithAlphabet(superuser.Collection().OTP.Length, "1234567890")

			otp := core.NewOTP(app)
			otp.SetCollectionRef(superuser.Collection().Id)
			otp.SetRecordRef(superuser.Id)
			otp.SetPassword(pass)

			err = app.Save(otp)
			if err != nil {
				return fmt.Errorf("failed to create OTP: %w", err)
			}

			color.New(color.BgGreen, color.FgBlack).Printf("Successfully created OTP for superuser %q:", superuser.Email())
			color.Green("\n├─ Id:    %s", otp.Id)
			color.Green("├─ Pass:  %s", pass)
			color.Green("└─ Valid: %ds\n\n", superuser.Collection().OTP.Duration)
			return nil
		},
	}

	return command
}
