package admin_console

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"

	"github.com/evgeniums/go-backend-helpers/pkg/admin"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/console_tool"
	"github.com/jessevdk/go-flags"
	"golang.org/x/term"
)

const AdministratorGroup string = "administrator"

type ManagerBuilder = func(app app_context.Context) *admin.Manager

func DefaultAdminManager(app app_context.Context) *admin.Manager {
	manager := admin.NewManager()
	manager.Init(app.Validator())
	return manager
}

type Administrator struct {
	MakeManager ManagerBuilder
}

func (a *Administrator) NewAdminManager(app app_context.Context, login ...string) *admin.Manager {
	manager := a.MakeManager(app)
	if len(login) > 0 {
		CheckAdminLogin(manager, login[0])
	}
	return manager
}

func (a *Administrator) Handlers(ctxBuilder console_tool.ContextBulder, parser *flags.Parser) {

	if a.MakeManager == nil {
		a.MakeManager = DefaultAdminManager
	}

	administrator, err := parser.AddCommand(AdministratorGroup, "Manage administrators", "", &console_tool.Dummy{})
	if err != nil {
		fmt.Printf("failed to add administrator group: %s", err)
		os.Exit(1)
	}

	adminLogin := AdminLogin{Administrator: a}
	addAdmin := AddAdmin{CtxBuilder: ctxBuilder, AdminLogin: adminLogin}

	administrator.AddCommand(AddAdminCmd, AddAdminDescription, "", &addAdmin)
	administrator.AddCommand(AdminPasswordCmd, AdminPasswordDescription, "", &AdminPassword{CtxBuilder: ctxBuilder, AdminLogin: adminLogin})
	administrator.AddCommand(AdminPhoneCmd, AdminPhoneDescription, "", &AdminPhone{CtxBuilder: ctxBuilder, AddAdmin: addAdmin})
	administrator.AddCommand(BlockAdminCmd, BlockAdminDescription, "", &BlockAdmin{CtxBuilder: ctxBuilder, AdminLogin: adminLogin})
	administrator.AddCommand(UnblockAdminCmd, UnblockAdminDescription, "", &UnblockAdmin{CtxBuilder: ctxBuilder, AdminLogin: adminLogin})
	administrator.AddCommand(ListAdminsCmd, ListAdminsDescription, "", &ListAdmins{CtxBuilder: ctxBuilder, Administrator: a})
}

type AdminLogin struct {
	Administrator *Administrator
	Login         string `long:"login" description:"Administrator login" required:"true"`
}

const AddAdminCmd string = "add"
const AddAdminDescription string = "Add administrator"

type AddAdmin struct {
	CtxBuilder console_tool.ContextBulder
	AdminLogin
	Phone string `long:"phone" description:"Administrator's phone number for SMS confirmations" required:"true"`
}

func CheckAdminLogin(manager *admin.Manager, login string) {
	err := manager.ValidateLogin(login)
	if err != nil {
		panic("invalid login format: must be alphanumeric or email in low case")
	}
}

func (a *AddAdmin) Execute(args []string) error {

	ctx := a.CtxBuilder(AdministratorGroup, AddAdminCmd)
	defer ctx.Close()
	manager := a.Administrator.NewAdminManager(ctx.App(), a.Login)

	fmt.Println("Please, enter new password:")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(fmt.Sprintf("failed to enter password: %s", err))
	}

	_, err = manager.AddAdmin(ctx, a.Login, string(password), a.Phone)
	return err
}

const AdminPasswordCmd string = "password"
const AdminPasswordDescription string = "Set new password for administrator"

type AdminPassword struct {
	CtxBuilder console_tool.ContextBulder
	AdminLogin
}

func (a *AdminPassword) Execute(args []string) error {

	ctx := a.CtxBuilder(AdministratorGroup, AdminPasswordCmd)
	defer ctx.Close()
	manager := a.Administrator.NewAdminManager(ctx.App(), a.Login)

	fmt.Println("Please, enter new password:")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(fmt.Sprintf("failed to enter password: %s", err))
	}

	return manager.SetPassword(ctx, a.Login, string(password), true)
}

const AdminPhoneCmd string = "phone"
const AdminPhoneDescription string = "Set new phone number for administrator"

type AdminPhone struct {
	CtxBuilder console_tool.ContextBulder
	AddAdmin
}

func (a *AdminPhone) Execute(args []string) error {
	ctx := a.CtxBuilder(AdministratorGroup, AdminPhoneCmd)
	defer ctx.Close()
	manager := a.Administrator.NewAdminManager(ctx.App(), a.Login)
	return manager.SetPhone(ctx, a.Login, a.Phone, true)
}

const BlockAdminCmd string = "block"
const BlockAdminDescription string = "Block administrator"

type BlockAdmin struct {
	CtxBuilder console_tool.ContextBulder
	AdminLogin
}

func (a *BlockAdmin) Execute(args []string) error {
	ctx := a.CtxBuilder(AdministratorGroup, BlockAdminCmd)
	defer ctx.Close()
	manager := a.Administrator.NewAdminManager(ctx.App(), a.Login)
	return manager.SetBlocked(ctx, a.Login, true, true)
}

const UnblockAdminCmd string = "unblock"
const UnblockAdminDescription string = "Unblock administrator"

type UnblockAdmin struct {
	CtxBuilder console_tool.ContextBulder
	AdminLogin
}

func (a *UnblockAdmin) Execute(args []string) error {
	ctx := a.CtxBuilder(AdministratorGroup, UnblockAdminCmd)
	defer ctx.Close()
	manager := a.Administrator.NewAdminManager(ctx.App(), a.Login)
	return manager.SetBlocked(ctx, a.Login, false, true)
}

const ListAdminsCmd string = "list"
const ListAdminsDescription string = "List administrators"

type ListAdmins struct {
	Administrator *Administrator
	CtxBuilder    console_tool.ContextBulder
}

func (a *ListAdmins) Execute(args []string) error {
	ctx := a.CtxBuilder(AdministratorGroup, ListAdminsCmd)
	defer ctx.Close()
	manager := a.Administrator.MakeManager(ctx.App())
	var admins []*admin.Admin
	err := manager.FindUsers(ctx, nil, &admins)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(admins, " ", " ")
	if err != nil {
		return fmt.Errorf("faild to serialize result: %s", err)
	}
	fmt.Printf("********************\n\n%s\n\n********************\n\n", string(b))
	return nil
}
