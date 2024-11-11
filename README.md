# Apparatus
Apparatus is a project that helps you getting a complete new development
environment running in minutes.

## Intro
- what is installed (asdf, butler, ..)
- how it is configured (default zsh etc..)

explain default configuration

## Installation
## OS
The Apparatus OS is based on [Fedora
Silberblue](https://fedoraproject.org/atomic-desktops/silverblue/) but does NOT
have it's own ISO which you can use to install. Instead you need to install
Silverblue first and then __rebase__ to Apparatus.

1. Install Fedora Silverblue using the ISO which you can find
   [here](https://fedoraproject.org/atomic-desktops/silverblue/download)
2. Once booted into Fedore Silverblue open a terminal and rebase to an unsigned
   image to get the proper signing keys and policies installed:
```
IMAGE_PATH=ghcr.io/vincentvdk/apparatus-os
IMAGE_TAG=latest
rpm-ostree rebase ostree-unverified-registry:$IMAGE_PATH:$IMAGE_TAG
systemctl reboot
````
3. Rebase to a signed image to complete the installation:
```
IMAGE_PATH=ghcr.io/vincentvdk/apparatus-os
IMAGE_TAG=latest
rpm-ostree rebase ostree-image-signed:docker://$IMAGE_PATH:$IMAGE_TAG
systemctl reboot
```

## Distrobox

## Details


## Fonts
To have the best experience use the patched version of your favorite font.
[](https://github.com/ryanoasis/nerd-fonts)

## Default tools and applications
- Neovim
- Zellij
- asdf
- chezmoi
- zsh

## ZSH
ZSH plugins are managed by [antidote](https://github.com/mattmc3/antidote). To
add or remove zsh plugins you can update the `zsh_plugins.txt` file which is
located in `files/bootstrap/zsh_plugins.txt` (used when building the image) or
`~/.config/zsh/zsh_plugins.txt` when using the distrobox created with this
container image

## Manage configuration
Initially some configuration (tmux, powerlevel10k, ..) was part of the
container image but as more configuration for other tools was added it made
more sense to keep everything in a git repository and use Chezmoi to sync.

Custom configuration of applications using `dotfiles` is managed by [chezmoi]()
which is already available in the image. The only pre requisit is that you
already manage your dotfiles with `chezmoi`

Check this link for more documentation
[https://www.chezmoi.io/user-guide/setup/](https://www.chezmoi.io/user-guide/setup/)

In a new Distrobox run:

```bash
$ chezmoi init --apply https://github.com/$GITHUB_USERNAME/dotfiles.git
```


## Distrobox
### Update a distrobox


## Shell  History


## ASDF

### Golang
Install golang plugin
```
asdf plugin add golang
```

Install a specific Golang version
```
asdf install golang 1.21.6
```

Configure __Global__ version
```
asdf global golang 1.21.6
```
> __NOTE__: if the global version is not matching with what was set, check the path in the output of `asdf current golang`. It is possible that the `.tool-versions` file is in the wrong `$HOME` dir. Simply delete the file and check again or run the global command again. 


Run `asdf reshim golang` after installing new packages
Switch between versions:
```
asdf shell golang 1.21.6
```
Run `asdf reshim golang` after installing new packages

Default go get packages
asdf-golang can automatically install a default set of packages with go get -u $PACKAGE right after installing a new Go version. To enable this feature, provide a $HOME/.default-golang-pkgs file that lists one package per line, for example:

```
// allows comments
github.com/Dreamacro/clash
github.com/jesseduffield/lazygit
```

You can specify a non-default location of this file by setting a ASDF_GOLANG_DEFAULT_PACKAGES_FILE variable.

## ENV VARS
