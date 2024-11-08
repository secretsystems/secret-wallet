package page_wallet

import (
	"encoding/json"
	"image"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/secretsystems/secret-wallet/animation"
	"github.com/secretsystems/secret-wallet/app_instance"
	"github.com/secretsystems/secret-wallet/components"
	"github.com/secretsystems/secret-wallet/containers/listselect_modal"
	"github.com/secretsystems/secret-wallet/containers/notification_modals"
	"github.com/secretsystems/secret-wallet/lang"
	"github.com/secretsystems/secret-wallet/prefabs"
	"github.com/secretsystems/secret-wallet/router"
	"github.com/secretsystems/secret-wallet/theme"
	"github.com/secretsystems/secret-wallet/utils"
	"github.com/secretsystems/secret-wallet/wallet_manager"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type PageContacts struct {
	isActive bool

	animationEnter *animation.Animation
	animationLeave *animation.Animation

	contactItems []*ContactListItem

	list              *widget.List
	buttonMenuContact *components.Button
}

var _ router.Page = &PageContacts{}

func NewPageContacts() *PageContacts {
	animationEnter := animation.NewAnimation(false, gween.NewSequence(
		gween.New(-1, 0, .25, ease.Linear),
	))

	animationLeave := animation.NewAnimation(false, gween.NewSequence(
		gween.New(0, -1, .25, ease.Linear),
	))

	list := new(widget.List)
	list.Axis = layout.Vertical

	menuIcon, _ := widget.NewIcon(icons.NavigationMenu)
	buttonMenuContact := components.NewButton(components.ButtonStyle{
		Icon:      menuIcon,
		Animation: components.NewButtonAnimationScale(.98),
	})

	return &PageContacts{
		animationEnter: animationEnter,
		animationLeave: animationLeave,

		list:              list,
		buttonMenuContact: buttonMenuContact,
	}
}

func (p *PageContacts) IsActive() bool {
	return p.isActive
}

func (p *PageContacts) Enter() {
	p.isActive = true
	page_instance.header.Title = func() string { return lang.Translate("Contacts") }
	page_instance.header.Subtitle = nil
	page_instance.header.ButtonRight = p.buttonMenuContact

	if !page_instance.header.IsHistory(PAGE_CONTACTS) {
		p.animationEnter.Start()
		p.animationLeave.Reset()
	}

	p.Load()
}

func (p *PageContacts) Leave() {
	p.animationLeave.Start()
	p.animationEnter.Reset()
}

func (p *PageContacts) Load() error {
	p.contactItems = make([]*ContactListItem, 0)

	wallet := wallet_manager.OpenedWallet
	contacts, err := wallet.GetContacts(wallet_manager.GetContactsParams{})
	if err != nil {
		return err
	}

	for _, contact := range contacts {
		item := NewContactListItem(contact)
		p.contactItems = append(p.contactItems, item)
	}

	return nil
}

func (p *PageContacts) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	{
		state := p.animationEnter.Update(gtx)
		if state.Active {
			defer animation.TransformX(gtx, state.Value).Push(gtx.Ops).Pop()
		}
	}

	{
		state := p.animationLeave.Update(gtx)
		if state.Active {
			defer animation.TransformX(gtx, state.Value).Push(gtx.Ops).Pop()
		}

		if state.Finished {
			p.isActive = false
			op.InvalidateOp{}.Add(gtx.Ops)
		}
	}

	if p.buttonMenuContact.Clicked() {
		go func() {

			addContactIcon, _ := widget.NewIcon(icons.SocialPersonAdd)
			downIcon, _ := widget.NewIcon(icons.FileFileDownload)
			upIcon, _ := widget.NewIcon(icons.FileFileUpload)

			keyChan := listselect_modal.Instance.Open([]*listselect_modal.SelectListItem{
				listselect_modal.NewSelectListItem("add_contact",
					listselect_modal.NewItemText(addContactIcon, lang.Translate("Add contact")).Layout,
				),
				listselect_modal.NewSelectListItem("import_contacts",
					listselect_modal.NewItemText(downIcon, lang.Translate("Import contacts")).Layout,
				),
				listselect_modal.NewSelectListItem("export_contacts",
					listselect_modal.NewItemText(upIcon, lang.Translate("Export contacts")).Layout,
				),
			})

			for key := range keyChan {
				switch key {
				case "add_contact":
					page_instance.pageContactForm.ClearForm()
					page_instance.pageRouter.SetCurrent(PAGE_CONTACT_FORM)
					page_instance.header.AddHistory(PAGE_CONTACT_FORM)
				case "export_contacts":
					exportContacts := func() error {
						file, err := app_instance.Explorer.CreateFile("contacts.json")
						if err != nil {
							return err
						}

						wallet := wallet_manager.OpenedWallet
						contacts, err := wallet.GetContacts(wallet_manager.GetContactsParams{})
						if err != nil {
							return err
						}

						data, err := json.MarshalIndent(contacts, "", " ")
						if err != nil {
							return err
						}

						_, err = file.Write(data)
						if err != nil {
							return err
						}
						defer file.Close()

						return nil
					}

					err := exportContacts()
					if err != nil {
						notification_modals.ErrorInstance.SetText(lang.Translate("Error"), err.Error())
						notification_modals.ErrorInstance.SetVisible(true, 0)
					} else {
						notification_modals.SuccessInstance.SetText(lang.Translate("Success"), lang.Translate("Contacts exported."))
						notification_modals.SuccessInstance.SetVisible(true, 0)
					}
				case "import_contacts":
					importContacts := func() error {
						file, err := app_instance.Explorer.ChooseFile(".json")
						if err != nil {
							return err
						}

						reader := utils.ReadCloser{ReadCloser: file}
						data, err := reader.ReadAll()
						if err != nil {
							return err
						}

						var contacts []wallet_manager.Contact
						err = json.Unmarshal(data, &contacts)
						if err != nil {
							return err
						}

						wallet := wallet_manager.OpenedWallet
						for _, contact := range contacts {
							err = wallet.StoreContact(contact)
							if err != nil {
								return err
							}
						}

						return nil
					}

					err := importContacts()
					if err != nil {
						notification_modals.ErrorInstance.SetText(lang.Translate("Error"), err.Error())
						notification_modals.ErrorInstance.SetVisible(true, 0)
					} else {
						p.Load()
						notification_modals.SuccessInstance.SetText(lang.Translate("Success"), lang.Translate("Contacts imported."))
						notification_modals.SuccessInstance.SetVisible(true, 0)
					}
				}
			}
		}()
	}

	widgets := []layout.ListElement{}

	if len(p.contactItems) == 0 {
		return layout.Inset{
			Left: unit.Dp(30), Right: unit.Dp(30),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th, unit.Sp(16), lang.Translate("You didn't add any contacts yet."))
			return lbl.Layout(gtx)
		})
	}

	for i := 0; i < len(p.contactItems); i++ {
		widgets = append(widgets, func(gtx layout.Context, index int) layout.Dimensions {
			return p.contactItems[index].Layout(gtx, th)
		})
	}

	widgets = append(widgets, func(gtx layout.Context, index int) layout.Dimensions {
		return layout.Spacer{Height: unit.Dp(30)}.Layout(gtx)
	})

	listStyle := material.List(th, p.list)
	listStyle.AnchorStrategy = material.Overlay

	return listStyle.Layout(gtx, len(widgets), func(gtx layout.Context, index int) layout.Dimensions {
		return widgets[index](gtx, index)
	})
}

type ContactListItem struct {
	contact        wallet_manager.Contact
	buttonSelect   *components.Button
	buttonEdit     *components.Button
	listItemSelect *prefabs.ListItemSelect
	clickable      *widget.Clickable
}

func NewContactListItem(contact wallet_manager.Contact) *ContactListItem {
	buttonSelect := components.NewButton(components.ButtonStyle{
		Rounded:  components.UniformRounded(unit.Dp(5)),
		TextSize: unit.Sp(14),
		Inset: layout.Inset{
			Top: unit.Dp(6), Bottom: unit.Dp(6),
			Left: unit.Dp(7), Right: unit.Dp(7),
		},
		Animation: components.NewButtonAnimationDefault(),
	})
	buttonSelect.Label.Alignment = text.Middle
	buttonSelect.Style.Font.Weight = font.Bold

	buttonEdit := components.NewButton(components.ButtonStyle{
		Rounded:  components.UniformRounded(unit.Dp(5)),
		TextSize: unit.Sp(14),
		Inset: layout.Inset{
			Top: unit.Dp(6), Bottom: unit.Dp(6),
			Left: unit.Dp(7), Right: unit.Dp(7),
		},
		Animation: components.NewButtonAnimationDefault(),
	})
	buttonEdit.Label.Alignment = text.Middle
	buttonEdit.Style.Font.Weight = font.Bold

	return &ContactListItem{
		contact:        contact,
		listItemSelect: prefabs.NewListItemSelect(),
		clickable:      new(widget.Clickable),
		buttonSelect:   buttonSelect,
		buttonEdit:     buttonEdit,
	}
}

func (item *ContactListItem) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if item.clickable.Hovered() {
		pointer.CursorPointer.Add(gtx.Ops)
	}

	if item.buttonEdit.Clicked() {
		page_instance.pageContactForm.contact = &item.contact
		page_instance.pageRouter.SetCurrent(PAGE_CONTACT_FORM)
		page_instance.header.AddHistory(PAGE_CONTACT_FORM)
	}

	if item.buttonSelect.Clicked() {
		txtWalletAddr := page_instance.pageSendForm.walletAddrInput.txtWalletAddr
		txtWalletAddr.SetValue(item.contact.Addr)
		page_instance.header.GoBack()
	}

	return layout.Inset{
		Top: unit.Dp(0), Bottom: unit.Dp(10),
		Right: unit.Dp(30), Left: unit.Dp(30),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		r := op.Record(gtx.Ops)
		dims := item.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			r := op.Record(gtx.Ops)
			dims := layout.Inset{
				Top: unit.Dp(10), Bottom: unit.Dp(10),
				Left: unit.Dp(15), Right: unit.Dp(15),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Label(th, unit.Sp(20), item.contact.Name)
								label.Font.Weight = font.Bold
								return label.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								addr := utils.ReduceAddr(item.contact.Addr)
								label := material.Label(th, unit.Sp(16), addr)
								label.Color = theme.Current.TextMuteColor
								return label.Layout(gtx)
							}),
						)
					}),
				)
			})
			c := r.Stop()

			if item.clickable.Hovered() {
				pointer.CursorPointer.Add(gtx.Ops)
				paint.FillShape(gtx.Ops, theme.Current.ListItemHoverBgColor,
					clip.UniformRRect(
						image.Rectangle{Max: image.Pt(dims.Size.X, dims.Size.Y)},
						gtx.Dp(10),
					).Op(gtx.Ops),
				)
			}

			if item.clickable.Clicked() {
				item.listItemSelect.Toggle()
			}

			layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				item.buttonSelect.Text = lang.Translate("Select")
				item.buttonSelect.Style.Colors = theme.Current.ButtonPrimaryColors
				item.buttonEdit.Text = lang.Translate("Edit")
				item.buttonEdit.Style.Colors = theme.Current.ButtonPrimaryColors
				return item.listItemSelect.Layout(gtx, th, item.buttonSelect, item.buttonEdit)
			})

			c.Add(gtx.Ops)
			return dims
		})
		c := r.Stop()

		paint.FillShape(gtx.Ops, theme.Current.ListBgColor,
			clip.RRect{
				Rect: image.Rectangle{Max: dims.Size},
				SE:   gtx.Dp(10), SW: gtx.Dp(10),
				NW: gtx.Dp(10), NE: gtx.Dp(10),
			}.Op(gtx.Ops))

		c.Add(gtx.Ops)

		return dims
	})
}
