package web

import (
	"encoding/json"
	"fmt"
	"github.com/yourusername/hogan-chain/internal/app"
	"github.com/yourusername/hogan-chain/internal/approval"
	"github.com/yourusername/hogan-chain/internal/identity"
	"github.com/yourusername/hogan-chain/pkg/l1_engine"
	"github.com/yourusername/hogan-chain/pkg/l2_business"
	"github.com/yourusername/hogan-chain/pkg/l3_domain"
	"github.com/yourusername/hogan-chain/pkg/l3_tenant"
	"github.com/yourusername/hogan-chain/pkg/l4_sandbox"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	app  *app.Application
	tmpl *template.Template
}

func New(a *app.Application) (*Server, error) {
	t, err := template.ParseFiles("index.html")
	if err != nil {
		return nil, err
	}
	return &Server{app: a, tmpl: t}, nil
}
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", s.page)
	mux.HandleFunc("/ui/overview", s.overview)
	mux.HandleFunc("/ui/role", s.setRole)
	mux.HandleFunc("/ui/action/", s.action)
	mux.HandleFunc("/api/state", s.apiState)
	return s.withIdentity(mux)
}
func (s *Server) withIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := "HoganChain_prime"
		if c, err := r.Cookie("hc_identity"); err == nil && c.Value != "" {
			id = c.Value
		}
		r.Header.Set("X-Identity", id)
		next.ServeHTTP(w, r)
	})
}
func actor(r *http.Request) string { return r.Header.Get("X-Identity") }
func (s *Server) page(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_ = s.tmpl.Execute(w, map[string]any{"Identity": actor(r)})
}
func (s *Server) setRole(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("identity")
	if _, err := s.app.Identity.Get(id); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "hc_identity", Value: id, Path: "/", SameSite: http.SameSiteLaxMode})
	r.Header.Set("X-Identity", id)
	s.overview(w, r)
}
func (s *Server) overview(w http.ResponseWriter, r *http.Request) {
	id := actor(r)
	u, _ := s.app.Identity.Get(id)
	users, _ := s.app.Identity.List()
	assets, _ := s.app.Assets.ListAssets()
	tasks, _ := s.app.L2.ListTasks()
	subs, _ := s.app.L2.ListSubsidiaries()
	programs, _ := s.app.L2.ListPrograms()
	domains, _ := s.app.Domains.List()
	tenants, _ := s.app.Tenants.ListTenants()
	contracts, _ := s.app.Tenants.ListContracts()
	projects, _ := s.app.Sandbox.ListProjects()
	props, _ := s.app.Approvals.List()
	inst, _ := s.app.Minting.List()
	fmt.Fprintf(w, `<section id="overview" class="workspace"><header class="workspace-head"><div><p class="eyebrow">Active identity</p><h2>%s</h2><p>%s</p></div><form hx-post="/ui/role" hx-target="#overview"><select name="identity" onchange="this.form.requestSubmit()">`, u.DisplayName, u.Role)
	for _, x := range users {
		sel := ""
		if x.ID == id {
			sel = " selected"
		}
		fmt.Fprintf(w, `<option value="%s"%s>%s · %s</option>`, x.ID, sel, x.ID, x.Role)
	}
	fmt.Fprint(w, `</select></form></header>`)
	hgk := s.app.HGK.Snapshot()
	hgxc := s.app.HGXC.Snapshot()
	fmt.Fprintf(w, `<div class="metrics"><article><span>L1 blocks</span><strong>%d</strong></article><article><span>HGK supply</span><strong>%d</strong></article><article><span>HGXC supply</span><strong>%d</strong></article><article><span>L1 assets</span><strong>%d</strong></article><article><span>Approvals</span><strong>%d</strong></article><article><span>L4 projects</span><strong>%d</strong></article></div>`, hgk.BlockHeight, hgk.TotalSupply, hgxc.TotalSupply, len(assets), len(props), len(projects))
	fmt.Fprint(w, `<div class="grid">`)
	s.renderActions(w, u)
	s.listCard(w, "Operations", []string{fmt.Sprintf("%d tasks", len(tasks)), fmt.Sprintf("%d programs", len(programs)), fmt.Sprintf("%d subsidiaries", len(subs)), fmt.Sprintf("%d domains", len(domains)), fmt.Sprintf("%d tenants / %d contracts", len(tenants), len(contracts))})
	s.listCard(w, "Authoritative state", []string{fmt.Sprintf("%d L1 assets", len(assets)), fmt.Sprintf("%d issued instruments", len(inst)), fmt.Sprintf("%d pending/history approvals", len(props)), fmt.Sprintf("Bridge: %d HGK locked", s.app.Bridge.Snapshot().LockedHGK)})
	fmt.Fprint(w, `</div></section>`)
}
func (s *Server) listCard(w http.ResponseWriter, title string, items []string) {
	fmt.Fprintf(w, `<article class="card"><h3>%s</h3><ul>`, title)
	for _, x := range items {
		fmt.Fprintf(w, "<li>%s</li>", x)
	}
	fmt.Fprint(w, "</ul></article>")
}
func (s *Server) renderActions(w http.ResponseWriter, u *identity.User) {
	fmt.Fprint(w, `<article class="card action-card"><h3>Quick actions</h3><div class="actions">`)
	switch u.Role {
	case identity.RolePrime:
		fmt.Fprint(w, buttons([][2]string{{"Mine L1 block", "mine"}, {"Register demo asset", "asset"}, {"Approve first proposal", "approve"}, {"Create snapshot", "snapshot"}, {"Create HGT director", "director"}}))
	case identity.RoleManager, identity.RoleSubsidiaryDirector, identity.RoleProgramManager:
		fmt.Fprint(w, buttons([][2]string{{"Create task", "task"}, {"Create subsidiary", "subsidiary"}, {"Create program", "program"}, {"Create domain", "domain"}, {"Register tenant + contract", "tenant"}, {"Create tester", "tester"}, {"Submit mint proposal", "mint_proposal"}}))
	default:
		fmt.Fprint(w, buttons([][2]string{{"Create test account", "test_account"}, {"Create dApp project", "test_project"}, {"Record test run", "test_run"}, {"Submit first project", "submit_project"}}))
	}
	fmt.Fprint(w, `</div><div id="flash"></div></article>`)
}
func buttons(xs [][2]string) string {
	var b strings.Builder
	for _, x := range xs {
		fmt.Fprintf(&b, `<button hx-post="/ui/action/%s" hx-target="#overview" hx-swap="outerHTML" hx-confirm="Run: %s?">%s</button>`, x[1], x[0], x[0])
	}
	return b.String()
}
func (s *Server) action(w http.ResponseWriter, r *http.Request) {
	a := actor(r)
	act := strings.TrimPrefix(r.URL.Path, "/ui/action/")
	var err error
	switch act {
	case "mine":
		err = s.app.Identity.Authorize(a, "*")
		if err == nil {
			_, err = s.app.HGK.MineBlock(a, 1)
		}
	case "snapshot":
		err = s.app.Identity.Authorize(a, "*")
		if err == nil {
			_, err = s.app.Snapshot(a)
		}
	case "asset":
		err = s.app.Identity.Authorize(a, "*")
		if err == nil {
			n := time.Now().Unix()
			err = s.app.Assets.RegisterAsset(a, l1_engine.RWAAssetRecord{AssetID: fmt.Sprintf("ASSET-%d", n), SPVID: fmt.Sprintf("SPV-%d", n), AssetClass: "DEMO_OCA", Jurisdiction: "PH", Currency: "USD", DeterminedValue: 100000000, RecognizedLiabilities: 10000000, AuthorizedTokenizationBps: 6000, ControllerAddress: a})
		}
	case "approve":
		err = s.app.Identity.Authorize(a, "*")
		if err == nil {
			ps, _ := s.app.Approvals.List()
			for _, p := range ps {
				if p.Status == approval.Submitted {
					err = s.app.Approvals.Decide(p.ID, a, true, "approved from dashboard")
					break
				}
			}
		}
	case "director":
		err = s.app.Identity.Authorize(a, "*")
		if err == nil {
			id := fmt.Sprintf("director_%d", time.Now().Unix())
			err = s.app.Identity.Create(a, identity.User{ID: id, DisplayName: "HGT Subsidiary Director", Role: identity.RoleSubsidiaryDirector})
		}
	case "task":
		err = s.app.Identity.Authorize(a, "task.*")
		if err == nil {
			id := fmt.Sprintf("TASK-%d", time.Now().Unix())
			err = s.app.L2.CreateTask(a, l2_business.Task{ID: id, Title: "New research task", Type: l2_business.AIInference, Priority: "NORMAL", Budget: 10000})
		}
	case "subsidiary":
		err = s.app.Identity.Authorize(a, "subsidiary.*")
		if err == nil {
			id := fmt.Sprintf("SUB-%d", time.Now().Unix())
			err = s.app.L2.CreateSubsidiary(a, l2_business.Subsidiary{ID: id, Name: "Half-Gallon Subsidiary", Purpose: "Special project development", Budget: 100000})
		}
	case "program":
		err = s.app.Identity.Authorize(a, "program.*")
		if err == nil {
			id := fmt.Sprintf("PGM-%d", time.Now().Unix())
			err = s.app.L2.CreateProgram(a, l2_business.Program{ID: id, Name: "Emergence Program", Runtime: l2_business.SREAnalysis, Budget: 50000})
		}
	case "domain":
		err = s.app.Identity.Authorize(a, "domain.*")
		if err == nil {
			id := fmt.Sprintf("DOM-%d", time.Now().Unix())
			err = s.app.Domains.Create(a, l3_domain.Domain{ID: id, Name: "Operating Domain", Type: "PROJECT"})
		}
	case "tenant":
		err = s.app.Identity.Authorize(a, "tenant.*")
		if err == nil {
			n := time.Now().Unix()
			tid := fmt.Sprintf("TEN-%d", n)
			err = s.app.Tenants.Register(a, l3_tenant.Tenant{ID: tid, Name: "Tenant", Controller: a})
			if err == nil {
				ds, _ := s.app.Domains.List()
				if len(ds) == 0 {
					err = s.app.Domains.Create(a, l3_domain.Domain{ID: fmt.Sprintf("DOM-%d", n), Name: "Tenant Domain", Type: "PROJECT"})
					ds, _ = s.app.Domains.List()
				}
				if err == nil {
					cid := fmt.Sprintf("CON-%d", n)
					err = s.app.Tenants.CreateContract(a, l3_tenant.Contract{ID: cid, TenantID: tid, DomainID: ds[0].ID, Type: "OPERATING", Rights: []string{"domain.read", "workflow.execute"}})
					if err == nil {
						err = s.app.Tenants.ActivateContract(a, cid)
					}
				}
			}
		}
	case "tester":
		err = s.app.Identity.Authorize(a, "user.delegate")
		if err == nil {
			id := fmt.Sprintf("test_%d", time.Now().Unix())
			err = s.app.Identity.Create(a, identity.User{ID: id, DisplayName: "New Tester", Role: identity.RoleTester, Scopes: []identity.Scope{{Type: "sandbox", ResourceID: "default"}}})
		}
	case "mint_proposal":
		err = s.app.Identity.Authorize(a, "proposal.create")
		if err == nil {
			id := fmt.Sprintf("MINT-%d", time.Now().Unix())
			err = s.app.Approvals.Create(approval.Proposal{ID: id, Type: "ASSET_BACKED_MINT", Title: "Asset-backed mint proposal", RequestedBy: a, RequiredRole: "HOGANCHAIN_PRIME", Payload: map[string]any{"requested_exposure_cents": 1000000}})
		}
	case "test_account":
		err = s.app.Identity.Authorize(a, "test_account.*")
		if err == nil {
			err = s.app.Sandbox.CreateAccount(a, l4_sandbox.TestAccount{ID: fmt.Sprintf("TA-%d", time.Now().Unix())})
		}
	case "test_project":
		err = s.app.Identity.Authorize(a, "sandbox.*")
		if err == nil {
			err = s.app.Sandbox.CreateProject(a, l4_sandbox.Project{ID: fmt.Sprintf("DAPP-%d", time.Now().Unix()), Name: "L4 dApp Experiment", Type: "DAPP", Version: "0.1"})
		}
	case "test_run":
		err = s.app.Identity.Authorize(a, "sandbox.*")
		if err == nil {
			ps, _ := s.app.Sandbox.ListProjects()
			if len(ps) == 0 {
				err = fmt.Errorf("create a project first")
			} else {
				err = s.app.Sandbox.RecordRun(a, l4_sandbox.TestRun{ID: fmt.Sprintf("RUN-%d", time.Now().UnixNano()), ProjectID: ps[0].ID, Scenario: "baseline", Expected: "pass", Actual: "pass", Passed: true})
			}
		}
	case "submit_project":
		err = s.app.Identity.Authorize(a, "proposal.submit_to_manager")
		if err == nil {
			ps, _ := s.app.Sandbox.ListProjects()
			if len(ps) == 0 {
				err = fmt.Errorf("create a project first")
			} else {
				err = s.app.Sandbox.SubmitProject(a, ps[0].ID)
			}
		}
	default:
		err = fmt.Errorf("unknown action")
	}
	if err != nil {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"notify":{"message":%q,"type":"error"}}`, err.Error()))
	}
	s.overview(w, r)
}
func (s *Server) apiState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"hgk": s.app.HGK.Snapshot(), "hgxc": s.app.HGXC.Snapshot(), "bridge": s.app.Bridge.Snapshot()})
}
func parseInt(v string) int64 { n, _ := strconv.ParseInt(v, 10, 64); return n }
