#!/usr/bin/env bash

github_token="$GITHUB_TOKEN"
if [[ -z "$github_token" ]]; then
  echo "Please provide a GitHub token. You can create one at:"
  echo "  https://github.com/settings/tokens/new?scopes=repo,read:user,user:email,write:packages"
  echo "If you don't want the value to be stored in a file, leave it empty and you will be prompted for it later."
  echo "GitHub token: "
  read -s github_token
fi

github_user_name="$GITHUB_USER_NAME"
if [[ -z "$github_user_name" ]]; then
  github_user_name="$(git config --global user.name)"
fi
if [[ -z "$github_user_name" ]]; then
  echo "Please provide a GitHub user name. You can also configure it with:"
  echo "  git config --global user.name \"Your Name\""
  echo "If you don't want the value to be stored in a file, leave it empty and you will be prompted for it later."
  echo "GitHub user name: "
  read github_user_name
fi

github_user_email="$GITHUB_USER_EMAIL"
if [[ -z "$github_user_email" ]]; then
  github_user_email="$(git config --global user.email)"
fi
if [[ -z "$github_user_email" ]]; then
  echo "Please provide a GitHub user email. You can also configure it with:"
  echo "  git config --global user.email \"Your Email\""
  echo "If you don't want the value to be stored in a file, leave it empty and you will be prompted for it later."
  echo "GitHub user email: "
  read github_user_email
fi

if [[ -z "$NO_GPG" ]]; then
  gpg_id="$GPG_ID"
  if [[ -z "$gpg_id" ]]; then
    gpg_id="$(git config --global user.signingkey)"
  fi
  if [[ -z "$gpg_id" ]]; then
    echo "Please provide a GPG ID. You can also configure it by following:"
    echo "  https://docs.github.com/en/authentication/managing-commit-signature-verification/telling-git-about-your-signing-key"
    echo "If you don't want the value to be stored in a file, leave it empty and you will be prompted for it later."
    echo "GPG ID: "
    read gpg_id
  fi
  if [[ -n "$gpg_id" ]]; then
    gpg_passphrase="$GPG_PASSPHRASE"
    if [[ -z "$gpg_passphrase" ]]; then
      echo "Please provide a GPG passphrase for the key $gpg_id."
      echo "If you don't want the value to be stored in a file, leave it empty and you will be prompted for it later."
      echo "GPG passphrase: "
      read -s gpg_passphrase
    fi
    if [[ -n "$gpg_passphrase" ]]; then
      gpg_key="$GPG_KEY"
      if [[ -z "$gpg_key" ]]; then
        gpg_key="$(gpg --armor --pinentry-mode=loopback --passphrase "$gpg_passphrase" --export-secret-key "$gpg_id" -w0 | base64 -w0)"
      fi
    fi
  fi
fi

if [[ -z "$NO_MATRIX" ]]; then
  matrix_url="$MATRIX_URL"
  if [[ -z "$matrix_url" ]]; then
    echo "Please provide a Matrix URL. For example: https://matrix-client.matrix.org/"
    echo "If you don't want the value to be stored in a file, leave it empty and you will be prompted for it later."
    echo "Matrix URL: "
    read matrix_url
  fi
  matrix_user="$MATRIX_USER"
  if [[ -z "$matrix_user" ]]; then
    echo "Please provide a Matrix username."
    echo "If you don't want the value to be stored in a file, leave it empty and you will be prompted for it later."
    echo "Matrix username: "
    read matrix_user
  fi

  matrix_password="$MATRIX_PASSWORD"
  if [[ -z "$matrix_password" ]]; then
    echo "Please provide a Matrix password."
    echo "If you don't want the value to be stored in a file, leave it empty and you will be prompted for it later."
    echo "Matrix password: "
    read -s matrix_password
  fi
fi

export GITHUB_TOKEN="$github_token"
export GITHUB_USER_NAME="$github_user_name"
export GITHUB_USER_EMAIL="$github_user_email"

export NO_GPG="$NO_GPG"
export GPG_ID="$gpg_id"
export GPG_PASSPHRASE="$gpg_passphrase"
export GPG_KEY="$gpg_key"

export NO_MATRIX="$NO_MATRIX"
export MATRIX_URL="$matrix_url"
export MATRIX_USER="$matrix_user"
export MATRIX_PASSWORD="$matrix_password"

cat .env.template | envsubst > .env
