
#!/bin/bash
export KUBECONFIG=${HOME}/Development/before-installer-testing/vsphere-ipi/auth/kubeconfig

make operator-image operator-push IMG=quay.io/jcallen/vcf-migration-operator:latest
make console-plugin-image console-plugin-push CONSOLE_PLUGIN_IMG=quay.io/jcallen/vcf-migration-console-plugin:latest

oc get nodes
read -p "Press Enter to continue..."


make deploy IMG=quay.io/jcallen/vcf-migration-operator:latest
make deploy-console-plugin CONSOLE_PLUGIN_IMG=quay.io/jcallen/vcf-migration-console-plugin:latest
