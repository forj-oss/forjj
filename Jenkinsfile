pipeline {
    agent any

    stages {
        stage('Build') {
            steps {
                sh('''ls -la
                source ./build-env.sh --jenkins-context $WORKSPACE
                build.sh''')
            }
        }
        stage('Tests') {
            steps {
                sh('''source ./build-env.sh --jenkins-context $WORKSPACE
                go test''')
            }
        }
        stage('Deploy') {
            when { branch 'master' }
            steps {
                echo 'Deploying...'
                sh('''source ./build-env.sh --jenkins-context $WORKSPACE''')
            }
        }
    }
}
