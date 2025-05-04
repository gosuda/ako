package main

import "os"

func init() {
	ciTemplates["jenkins"] = createJenkinsConfig
}

const (
	jenkinsFileName = "Jenkinsfile"
	jenkinsTemplate = `pipeline {
    agent { 
        docker { image 'golang:1' }
    } 
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        stage('Lint') {
            steps {
                sh 'go tool golangci-lint run'
            }
        }
        stage('Test') {
            steps {
                sh 'go test ./...'
            }
        }
    }
}`
)

func createJenkinsConfig(name string) error {
	f, err := os.Create(jenkinsFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(jenkinsTemplate); err != nil {
		return err
	}

	if err := f.Sync(); err != nil {
		return err
	}

	return nil
}
