// Copyright 2017 The kubecfg authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/ksonnet/ksonnet/metadata"
	"github.com/ksonnet/ksonnet/pkg/kubecfg"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(initCmd)
	// TODO: We need to make this default to checking the `kubeconfig` file.
	initCmd.PersistentFlags().String(flagAPISpec, "version:v1.7.0",
		"Manually specify API version from OpenAPI schema, cluster, or Kubernetes version")

	bindClientGoFlags(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init <app-name>",
	Short: "Initialize a ksonnet application",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		if len(args) != 1 {
			return fmt.Errorf("'init' takes a single argument that names the application we're initializing")
		}

		appName := args[0]
		appDir, err := os.Getwd()
		if err != nil {
			return err
		}
		appRoot := metadata.AbsPath(path.Join(appDir, appName))

		specFlag, err := flags.GetString(flagAPISpec)
		if err != nil {
			return err
		}

		context, err := flags.GetString(flagEnvContext)
		if err != nil {
			return err
		}

		log.Infof("Creating a new app '%s' at path '%s'", appName, appRoot)

		// Find the URI and namespace of the current cluster, if it exists.
		var ctx *string
		if len(context) != 0 {
			ctx = &context
		}
		uri, namespace, err := resolveContext(ctx)
		if err != nil {
			return err
		}

		c, err := kubecfg.NewInitCmd(appName, appRoot, specFlag, &uri, &namespace)
		if err != nil {
			return err
		}

		return c.Run()
	},
	Long: `Initialize a ksonnet application in a new directory,` + " `app-name`" + `.

This command generates all the project scaffolding required to begin creating and
deploying components to Kubernetes clusters.

ksonnet applications are initialized based on your current cluster configurations,
as defined in your` + " `$KUBECONFIG` " + `environment variable. The *Examples*
section below demonstrates how to customize your configurations.

Creating a ksonnet application results in the following directory tree.

    app-name/
      .ksonnet/      Metadata for ksonnet
      app.yaml       Application specifications, ex: name, api version
      components/    Top-level Kubernetes objects defining the application
      environments/  Kubernetes cluster definitions
        default/     Default environment, initialized from the current kubeconfig
          .metadata/ Contains a versioned ksonnet-lib, see [1] for details
      lib/           user-written .libsonnet files
      vendor/        part libraries, prototypes

To begin populating your ksonnet application, see the docs for` + " `ks generate` " + `.

[1] Each environment uses a specific version of ksonnet-lib. Users can set flags
to generate the library based on a variety of data, including server
configuration and an OpenAPI specification of a Kubernetes build. By default,
this is generated from the capabilities of the cluster specified in the cluster
of the current context specified in` + " `$KUBECONFIG`" + `.
`,
	Example: `# Initialize a ksonnet application, based on cluster information from the
# active kubeconfig file (specified by the environment variable $KUBECONFIG).
# More specifically, the current context is used.
ks init app-name

# Initialize a ksonnet application, using the context 'dev' from the current
# kubeconfig file ($KUBECONFIG). This initializes the default environment
# using the server address and default namespace located at the context 'dev'.
ks init app-name --context=dev

# Initialize a ksonnet application, using the context 'dev' and the namespace
# 'dc-west' from the current kubeconfig file ($KUBECONFIG). This initializes
# the default environment using the server address at the context 'dev' for
# the namespace 'dc-west'.
ks init app-name --context=dev --namespace=dc-west

# Initialize a ksonnet application, using v1.7.1 of the Kubernetes OpenAPI spec
# to generate 'ksonnet-lib'.
ks init app-name --api-spec=version:v1.7.1

# Initialize a ksonnet application, using the OpenAPI spec generated by a
# specific build of Kubernetes to generate 'ksonnet-lib'.
ks init app-name --api-spec=file:swagger.json`,
}
