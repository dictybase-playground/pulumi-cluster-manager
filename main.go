package main

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		appName := "modware-import"
		appLabels := pulumi.StringMap{
			"app": pulumi.String(appName),
		}
		commands := []string{"/usr/local/bin/content"}
		args := []string{"content-data"}
		envVars := corev1.EnvVarArray{
			corev1.EnvVarArgs{
				Name:  pulumi.String("S3_BUCKET_PATH"),
				Value: pulumi.String("ADD_S3_BUCKET_PATH_VALUE"),
			},
		}
		deployment, err := appsv1.NewDeployment(ctx, appName, &appsv1.DeploymentArgs{
			Spec: appsv1.DeploymentSpecArgs{
				Selector: &metav1.LabelSelectorArgs{
					MatchLabels: appLabels,
				},
				Replicas: pulumi.Int(1),
				Template: &corev1.PodTemplateSpecArgs{
					Metadata: &metav1.ObjectMetaArgs{
						Labels: appLabels,
					},
					Spec: &corev1.PodSpecArgs{
						Containers: corev1.ContainerArray{
							corev1.ContainerArgs{
								Name:    pulumi.String(appName),
								Image:   pulumi.String("dictybase/modware-import:sha-02a6dcd"),
								Command: pulumi.ToStringArray(commands),
								Args:    pulumi.ToStringArray(args),
								Env:     envVars,
							},
						},
					},
				},
			},
		})
		if err != nil {
			return err
		}

		ctx.Export("name", deployment.Metadata.Name())
		return nil
	})
}
