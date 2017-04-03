# Introduction

Ceci est un petit document de design pour mettre en oeuvre le Forjfile

# Details

Pour faire simple, tous les paramètres que l'on pourrait passer sur un
`forjj create` ou `forjj update` peut être remplacé par un `Forjfile`.

Les paramètres du Wokspace devrait être conservé.
Par contre, repo/app/autres ne devrait plus l'être à priori.

`forjj maintain` n'est pas concerné. Ce dernier se base sur le repo.

## Comment forjj create marche?

```bash
$ vim Forjfile
$ forjj create
```

C'est le cas le plus simple.

Il lit le fichier et positionne toutes les données pour créer la forge.
Ensuite, il la crée et copie une version sans les données Workspace mais
avec les données infra (nom de l'organisation, nom du repo infra)

Il est possible aussi d'y ajouter des options qui pourrait être positionné
dans le fichier, pour le rendre plus générique.

On pourra probablement ajouter d'autre Forjfile avec une option -f afin
de les merger et utiliser la résultante.
Mais ce sera une étape ultérieure.

Le contenu du Forjfile est défini par l'ensemble des objets cli que Forjj
manipule via cli.

Peut-être pourrions-nous implémenter cette mécanique dans cli ou de Forjj
Si, on l'implémente dans cli, on exploite toutes les finesses de cli et
cela necessite de définir un tag yaml si besoin
Le fichier pourrait être géré à plat.
Ex:

```yaml
---
workspace:
  docker-exe-path: "blabla"
  [...]
apps/repos/...:
  <instance>:
    key: 'val'
```

L'avantage, c'est la lisibilité.
On pourrait définir workspace par un nom générique (tag) `forj`

```yaml
---
forj:
  docker-exe-path: "blabla"
  [...]
apps/repos/...:
  <instance>:
    key: 'val'
```

Concernant `infra`, le nom de l'organisation ou de repo infra peuvent être
défini, car ils font partie de l'objet.
On peut garder le nom tel quel...

Voici, un autre exemple:
```yaml
local: # workspace object - Define the local path
  forjj-contribs-path: /home/.../forjj-contribs
infra: # infra object - Define the infra repository information.
  name: myinfra_name
  remote-infra-url: git@github.com/...
  flow: myflow
forj: # Defaults for the forge
  organization: myorganizationName

```

Voila l'exemple le plus simple pour démarrer une forge à partir de rien.
Le minimum requis: `Forjfile`
```yaml
forj:
  organization: myOrg
apps:
  jenkins:
    type: ci
  github:
    type: upstream
```

A coté de cela, on a besoin des données d'identification pour github
fichier : `forj-creds.yml`
```yaml
apps:
  github:
    token: blabla
```

Si je veux ces données dans un système sécurisé, comme vault, on pourra le
définir dans fichier Forjfile et le creds défini le token vault par ex.

Ce dernier exemple prédefini plusieurs chose de base:
- github est un service externe défini.
- jenkins est un nouveau service démarré par docker.
  Le nom de la machine définira la public url.
- tout le reste est du source généré.

Au vu de l'exemple précédent, j'ai besoin de définir des infos principales
que définiront la forj, comme flow ou l'organization name

Du coup, je pense que forjj doit intégrer les objets suivants:
- workspace (tag local) - Definit les pointeurs locaux comme des paths.
- infra - Information concernant l'infra (repo name, etc...)
- forj - Les paramètres par défaut de la forge
- app - Liste d'applications.
- repo - Liste de sources repos.

Si je penche pour l'intégration dans forjj plutôt que cli, je vais devoir
faire une fonction à part qui lirait le fichier, mais en déterminant
une liste d'objet dépendant du contenu. Donc cela se fera après la lecture
des drivers.

Dans cli, cette étape est aussi obligatoire. Du coup c'est pareil de ce
point de vue...

par contre, l'analyse du fichier Forjfile pourrait être plus simple.
On lit en générique via map[]...

On fait la distinction des objets multi-instance contre les mono-instance
Du coup, le fichier pour être simple à lire, il faudrait 2 structures
différente.
Une structure mono et une multi.

Comment un driver positionnant un object mono devrait permettre la définition
dans Forjfile?

De plus, il est possible que le driver définisse cela uniquement au niveau
du driver, mais pas au niveau de Forjj.

Le cas driver pourrait se simplifier au nom du driver dans l'instance_name

es-ce que le cas applicatif devrait être remonté au niveau Forj?

Comment simplifier la lisibilité et la programmation en même temps?

### Derrière la scène en go - Forjfile

il faut quelque chose comme ca pour la lecture.
```golang
type ForjfileCore map[string]map[string]interface{}
```

la partie interface pourrait être `string` ou `map[string]string`

Je pense qu'en ecriture, ca devrait marcher... A vérifier.

Si la lecture/l'ecriture marche, je vais pouvoir copier le contenu du Forjfile
 au bon endroit si nécessaire. Mais également, l'enrichir via les
 options des objets qui sont donnés à `forjj`.


# Comment ca s'implémente?

Il faut les fonctions qui lit/écrit le fichier Forjfile.

Ce dernier peut-être dans le répertoire courant
Si, le rep courant (ou --workspace) est un workspace, on lit Forjfile
et on ignore la partie workspace si elle existe, car cette dernière
sera chargé via le workspace.

es-ce que je définis la structure yaml d'un Forjfile pour l'utiliser
et l'écrire?

le lien avec cli pourrait être simplifié si je défini ou les données des
flags seront stockés...


On pourrait avoir ceci:
```golang
type Forjfile struct {
    Forj ForjDefault
    Infra map[string]string
    Workspace map[string]string `yaml:"local"`
    Singles map[string]map[string]string `yaml:,inline`
    Repos map[string]map[string]string
    Apps map[string]map[string]string
    Instances map[string]map[string]map[string]string `yaml:,inline`
}

type ForjDefault struct {
    InfraName string
    Upstream string
    More map[string]string `yaml:,inline`
}
```

Ca pourrait marcher. Mais j'ai un sérieux doute sur la lecture yaml.
L'avantage de cette structure c'est que le code est plus explicite
ex:
`value := forjfile.Forj["test"]`

Si j'utilise cli:
`value := forjfile["forj"]["test"]`

L'idéal, c'est
`value := forjfile.Forj.test`

et les données définis par les drivers:

`value := forjfile.Forj.More["plugin_par"]`

L'écriture est facile.

