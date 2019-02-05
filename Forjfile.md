# Forjfile

Forjfile is the core source file to build a DevOps environment.

Here is a typical simple Forjfile.

```yaml
applications:
  github:
    type: upstream
  jenkins:
    type: ci
infra:
  organization: my-organization
repositories:
  example:
    title: Repository of examples
```

Here is a complete Forjfile description.

Note that each plugins may support :
- additional options, to add to application objects. (/applications/...)
- Additional objects. Ex: `/projects` is defined by `jenkins` plugin.


```yaml
local-settings:
  docker-exe-path: "~/tmp/docker-1.12.1"
  contribs-repo: "~/src/forj/go_workspace/src/github.com/forjj-contribs"
  flows-repo: "~/src/forj/go_workspace/src/github.com/forjj-flows"
  repotemplates-repo: "~/src/forj/go_workspace/src/github.com/forjj-repotemplates"
# Ces settings sont remplis au moment de l'instanciation d'une forj. Elle demande le nom de l'organisation ou --organization ou défini par le workspace path.
infra: # Object infra
  # Peut être fixé en dur ici. Sinon, vient du --organization ou du nom du répertoire root du workspace.
  organization: "forjj-test"
  # Peut être fixé en dur. Mais en général, il est ajouté dans une copie du fichier posée dans le repo. Il peut-être à "none"
  # Si forj voit cette copie dans le répertoire courant, par défaut, il va chercher le workspace. S'il ne le trouve pas, il râle et réclame le workspace --workspace ou cd dans le bon répertoire.
  remote: git@github.com:forjj-test/forjj-test-infra.git
applications: # Object app
  github: # C'est le nom de l'instance.
    # driver: github - optionnel si le nom de l'instance est le même que le nom du driver.
    type: upstream
    # par défaut les données secure sont dans un fichier creds.yaml dans le workspace. On pourra changer cela via :
    # secure-token: scheme://<service>/<path/key>
    # Ex: Vault pourrait être utilisé via un token Vault à fournir (creds? Env? autre?)
    # secure-token: vault://localhost:1234/path/key
  jenkins:
    type: ci
    # jobdsl-repo: "" Par défaut, c'est le répo infra. Il faut que le plugin demande à forjj un montage de ce repo et ou il est monté.
    # jobdsl-path: "jobs-dsl" C'est la chaine par défaut. Il sera ajouté avec /* dans le seedjob.
    dockerfile-from-image: hub.docker.hpecorp.net/devops/jenkins-dood
repositories:
  example:
    flow: "pull_request" # Optionnel: Par défaut, pas de flow.
    title: "Repo de test"
    # upstream: github // Optionnel si le défaut est défini ou s'il n'y en a qu'un seul.
  report-api:
    # En précisant l'upstream; forjj fera un checkout de ce dernier, même s'il n'est pas de la même organization.
    # Il ne sera pas créé, vide ou pré-rempli (repotemplates)
    # Il sera juste identifié par la forge pour travailler avec notamment quand les flows créeront un objet project sur jenkins avec les infos du repo.
    upstream: https://github.hpe.com/change-records/report-api
forj:
  organization: "ChangeRecords"
  users: # Objet user créé par le driver github
    christophe-larsonneur:
      email: christophe.larsonneur@hpe.com
  groups: # Objet créé par le driver github
    forj-core: # On pourra y faire référence dans la définition d'un repo.
      members: [ "christophe-larsonneur", "miguel-quintero" ]
  project: # C'est un objet créé par jenkins pour créer des jobs-dsl
    report-api: # Ici, on va créer un jobdsl pour report-api.
      upstream: https://github.hpe.com/change-records/report-api
```
