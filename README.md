### Nutri price server
A Go-based REST API service for managing and calculating data related to products, their nutritional value, and meals. The service supports:

* CRUD Operations: Manage purchased products, product nutritional values, and recipes.
* Meal Calculations: Compute meal prices and nutritional values based on selected products and recipes.

### Installation

Prerequisites:
* Go (version 1.23.1 used during development)
* Docker (version 27.3.1 used during development)
* just (optional, used to run commands using just)
* The environment variables mentioned below must be set. Copy `.env.example` file and rename it to `.env`.

To run the app, execute the command `just start`. Check the `justfile`for other available commands.

### Environment variables

* DATABASE_URL - url to connect to postgres database
* PORT - port to run the API server

### Development Notes

#### Enable autocompletion for justfile commands
* Generate completion script
    * just --completions bash > ~/.just-completions.bash
* Add following line to ~/.bashrc
    * source ~/.just-completions.bash
* Reload shell
    * source ~/.bashrc

#### Install gcloud
https://cloud.google.com/sdk/docs/install#deb
```
sudo apt-get update

sudo apt-get install apt-transport-https ca-certificates gnupg curl

curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg

echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list

sudo apt-get update && sudo apt-get install google-cloud-cli

gcloud init
```

Other useful commands
* gcloud auth login
* gcloud config set project PROJECT_ID
* gcloud auth configure-docker
    * run before pushing docker image first time after gcloud installation