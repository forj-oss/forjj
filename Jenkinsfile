pipeline {
    agent any

    stages {
        stage('Build') {
            steps {
                sh('''source build-env $WORKSPACE
                build.sh''')
            }
        }
        stage('Tests') {
            steps {
                sh('''source build-env $WORKSPACE
                go test''')
            }
        }
        stage('Deploy') {
            when { branch 'master' }
            steps {
                echo 'Deploying....'
            }
        }
    }
}
