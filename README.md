# civoctl

CivoCtl is a simple controller to reconcile a list of cluster names with actual k3s clusters in Civo cloud.

CivoCtl accepts a yaml list of clusters. As cluster names are added or deleted from the lists, civoctl will create and delete clusters through the Civo API

```yaml
clusters:
  - name: civo-1
    nodes: 3
  - name: civo-2
    nodes: 2
```

_Note_: Civoctl will watch the file for changes, any updates made to the configuration file will be realized during the subsequent reconcilliation cycle.

## Usage

### Download

```bash
go get -u github.com/gabeduke/civoctl
```

### Configure

I recommend starting with a list of clusters that already exist in your account. The create loop will be a noop and the delete will not be triggered because the clusters exist.

Civoctl will search in the following places for a config file (named `.civoctl.yaml`)

```bash
# etc
/etc/civoctl/.civoctl.yaml

# home
$HOME/.civoctl.yaml

# current working dir
./.civoctl.yaml
```

Go ahead and stamp the cfg file:

```bash
cat <<EOF > .civoctl.yaml
clusters:
  - name: civo-1
  - name: civo-2
EOF
```

### Run

```bash
# export Civo API token (this can also be passed as the cmd line flag --token)
# export CIVO_TOKEN=$(cat /var/secrets/civo_token)

# Show Civoctl help
civoctl --help

# Run the control loop to create clusters
civoctl run

# Run in dangerous mode to also delete clusters
# civoctl run -d
```