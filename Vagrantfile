# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|
  config.vm.provider :vmware_desktop do |vmware|
    vmware.gui = false
    vmware.vmx["memsize"] = "1024"
    vmware.vmx["numvcpus"] = "1"
    vmware.vmx["vhv.enable"] = "TRUE"
  end

  config.vm.define "buildkite" do |buildkite|
    buildkite.vm.network :private_network, ip: "192.168.60.150"
    buildkite.vm.box = "bento/ubuntu-19.04"
    buildkite.vm.hostname = "buildkite"
    buildkite.vm.network "private_network", type: "dhcp"

    buildkite.vm.post_up_message = <<~EOF
    👋🏼 Hello World!
    EOF
  end

  config.vm.synced_folder ".", "/vagrant", disabled: true

  config.vm.provision "base", type: "shell", inline: <<~BASE, privileged: false
  sudo apt-get update
  sudo DEBIAN_FRONTEND=noninteractive apt-get install -y git curl
  curl -sSL https://dl.google.com/go/go1.12.6.linux-amd64.tar.gz -o go1.12.6.linux-amd64.tar.gz
  sudo tar -xvf go1.12.6.linux-amd64.tar.gz
  sudo mv go /usr/local
  BASE
end
