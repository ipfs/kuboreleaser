#!/usr/bin/env bash

github_token="$GITHUB_TOKEN"
if [[ -z "$github_token" ]]; then
  echo "Please provide a GitHub token. You can create one at:"
  echo "  https://github.com/settings/tokens/new?scopes=repo,read:user,user:email,write:packages"
  echo "GitHub token: "
  read -s github_token
fi
if [[ -z "$github_token" ]]; then
  echo "GitHub token is required"
  exit 1
fi

github_user_name="$GITHUB_USER_NAME"
if [[ -z "$github_user_name" ]]; then
  github_user_name="$(git config --global user.name)"
fi
if [[ -z "$github_user_name" ]]; then
  echo "Please provide a GitHub user name. You can also configure it with:"
  echo "  git config --global user.name \"Your Name\""
  echo "GitHub user name: "
  read github_user_name
fi
if [[ -z "$github_user_name" ]]; then
  echo "GitHub user name is required"
  exit 1
fi

github_user_email="$GITHUB_USER_EMAIL"
if [[ -z "$github_user_email" ]]; then
  github_user_email="$(git config --global user.email)"
fi
if [[ -z "$github_user_email" ]]; then
  echo "Please provide a GitHub user email. You can also configure it with:"
  echo "  git config --global user.email \"Your Email\""
  echo "GitHub user email: "
  read github_user_email
fi
if [[ -z "$github_user_email" ]]; then
  echo "GitHub user email is required"
  exit 1
fi

if [[ -z "$NO_GPG" ]]; then
  gpg_id="$GPG_ID"
  if [[ -z "$gpg_id" ]]; then
    gpg_id="$(git config --global user.signingkey)"
  fi
  if [[ -z "$gpg_id" ]]; then
    echo "Please provide a GPG ID. You can also configure it by following:"
    echo "  https://docs.github.com/en/authentication/managing-commit-signature-verification/telling-git-about-your-signing-key"
    echo "GPG ID: "
    read gpg_id
  fi
  if [[ -z "$gpg_id" ]]; then
    echo "GPG ID is required"
    exit 1
  fi

  gpg_passphrase="$GPG_PASSPHRASE"
  if [[ -z "$gpg_passphrase" ]]; then
    echo "Please provide a GPG passphrase for the key $gpg_id."
    echo "GPG passphrase: "
    read -s gpg_passphrase
  fi
  if [[ -z "$gpg_passphrase" ]]; then
    echo "GPG passphrase is required"
    exit 1
  fi

  gpg_key="$GPG_KEY"
  if [[ -z "$gpg_key" ]]; then
    gpg_key="$(gpg --armor --pinentry-mode=loopback --passphrase "$gpg_passphrase" --export-secret-key "$gpg_id" -w0 | base64 -w0)"
  fi
  if [[ -z "$gpg_key" ]]; then
    echo "GPG key is required"
    exit 1
  fi
fi

matrix_user="$MATRIX_USER"
if [[ -z "$matrix_user" ]]; then
  echo "Matrix username: "
  read matrix_user
fi
if [[ -z "$matrix_user" ]]; then
  echo "Matrix username is required"
  exit 1
fi

matrix_password="$MATRIX_PASSWORD"
if [[ -z "$matrix_password" ]]; then
  echo "Matrix password: "
  read -s matrix_password
fi
if [[ -z "$matrix_password" ]]; then
  echo "Matrix password is required"
  exit 1
fi

export GITHUB_TOKEN="$github_token"
export GITHUB_USER_NAME="$github_user_name"
export GITHUB_USER_EMAIL="$github_user_email"

export NO_GPG="$NO_GPG"
export GPG_ID="$gpg_id"
export GPG_PASSPHRASE="$gpg_passphrase"
export GPG_KEY="$gpg_key"

export MATRIX_USER="$matrix_user"
export MATRIX_PASSWORD="$matrix_password"

cat .env.template | envsubst > .env
