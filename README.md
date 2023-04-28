# Ansible Script for Kafka and Zookeeper

This Ansible script is designed to automate the installation and configuration of Kafka and Zookeeper on multiple servers. It also includes a way to set up the Kafka cli, kaf, which provides an easy-to-use interface for managing Kafka clusters.

## Prerequisites

Before running this script, make sure you have the following prerequisites:

- Ansible, python installed on your local, server machine
- SSH access to the target servers

## Usage

- Clone the repository:

```bash
git clone https://github.com/joeirimpan/kaf-ansible.git
```

- Change `kafka_version` under `group_vars/kafka_nodes` to setup a different kafka version.
- Update the `group_vars/kafka_nodes` file with the appropriate configuration settings for your Kafka and Zookeeper installation.
- Open hosts.sample inventory file and change the addresses under `[kafka_nodes]` section
- Run the Ansible playbook

```bash
cp hosts.sample hosts
ansible-playbook -i hosts tasks/dependencies.yaml
ansible-playbook -i hosts tasks/kaf.yaml
ansible-playbook -i hosts tasks/main.yaml
```

The playbook will install and configure Kafka and Zookeeper on the target servers based on the settings defined in the group_vars/kafka_nodes file. It will also install and configure kaf, the Kafka cli.

- Run kaf cli to verify

```bash
kaf topics
kaf topic create test
```

## Generating configs for multiple servers

```bash
go build -o byok main.go
./byok --in config-in.yaml --addrs="0.0.0.0,0.0.0.1,0.0.0.2" --out config-out
```
Above command generates 3 config files with different broker ids, addresses. Change `var_files` in `tasks/main.yaml` to use appropriate config file for setting up the nodes. 

## Credits

https://github.com/sleighzy/ansible-kafka

## License

This Ansible script is released under the MIT License. Feel free to modify and use it as you see fit!
