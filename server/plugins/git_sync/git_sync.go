package git_sync

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"

	"github.com/checkr/codeflow/server/agent"
	"github.com/checkr/codeflow/server/plugins"
	git "github.com/libgit2/git2go"
)

type GitSync struct {
	events        chan agent.Event
	workdir       string
	rsaPrivateKey string
	rsaPublicKey  string
}

func init() {
	agent.RegisterPlugin("git_sync", func() agent.Plugin {
		return &GitSync{}
	})
}

func (x *GitSync) Description() string {
	return "Sync Git repositories and create new features"
}

func (x *GitSync) SampleConfig() string {
	return ` `
}

func (x *GitSync) Start(e chan agent.Event) error {
	x.events = e
	log.Println("Started GitSync")

	return nil
}

func (x *GitSync) Stop() {
	log.Println("Stopping GitSync")
}

func (x *GitSync) Subscribe() []string {
	return []string{
		"plugins.GitPing",
		"plugins.GitSync:update",
	}
}

func (x *GitSync) Process(e agent.Event) error {
	log.Printf("Process GitSync event: %s", e.Name)

	var err error

	gitSyncEvent := e.Payload.(plugins.GitSync)
	gitSyncEvent.Action = plugins.Status
	gitSyncEvent.State = plugins.Running
	gitSyncEvent.StateMessage = ""

	x.workdir = viper.GetString("plugins.git_sync.workdir")
	x.rsaPrivateKey = gitSyncEvent.Git.RsaPrivateKey
	x.rsaPublicKey = gitSyncEvent.Git.RsaPublicKey

	gitSyncEvent.State = plugins.Fetching
	err = x.fetchCode(&gitSyncEvent)
	if err != nil {
		gitSyncEvent.State = plugins.Failed
		gitSyncEvent.StateMessage = fmt.Sprintf("%v (Action: %v)", err.Error(), gitSyncEvent.State)
		event := e.NewEvent(gitSyncEvent, err)
		x.events <- event
		log.Println(err)
		return err
	}

	return nil
}

func (x *GitSync) fetchCode(gitSync *plugins.GitSync) error {
	repoPath := fmt.Sprintf("%s/%s", x.workdir, gitSync.Project.Repository)

	repo, err := x.findOrClone(repoPath, gitSync.Git.SshUrl, gitSync.Git.Public)
	if err != nil {
		return err
	}

	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		return err
	}

	fetchOptions := &git.FetchOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      x.credentialsCallback,
			CertificateCheckCallback: x.certificateCheckCallback,
		},
	}

	err = remote.Fetch([]string{}, fetchOptions, "")
	if err != nil {
		return err
	}

	return nil
}

func (x *GitSync) findOrClone(path string, cloneUrl string, public bool) (*git.Repository, error) {
	var repo *git.Repository
	var err error

	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			cloneOptions := &git.CloneOptions{}

			cloneOptions.FetchOptions = &git.FetchOptions{
				RemoteCallbacks: git.RemoteCallbacks{
					CredentialsCallback:      x.credentialsCallback,
					CertificateCheckCallback: x.certificateCheckCallback,
				},
			}
			cloneOptions.CheckoutOpts = &git.CheckoutOpts{Strategy: 1}

			repo, err = git.Clone(cloneUrl, path, cloneOptions)
		} else {
			return &git.Repository{}, err
		}
	} else {
		repo, err = git.OpenRepository(path)
	}

	return repo, err
}

func (x *GitSync) credentialsCallback(url string, username string, allowedTypes git.CredType) (git.ErrorCode, *git.Cred) {
	id_rsa_priv := x.rsaPrivateKey
	id_rsa_pub := x.rsaPublicKey
	ret, cred := git.NewCredSshKeyFromMemory("git", id_rsa_pub, id_rsa_priv, "")
	return git.ErrorCode(ret), &cred
}

// Made this one just return 0 during troubleshooting...
func (x *GitSync) certificateCheckCallback(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
	return 0
}
