package main

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		appName := "modware-import"
		appLabels := pulumi.StringMap{
			"app": pulumi.String(appName),
		}
		cfg := config.New(ctx, "")
		accessToken := cfg.RequireSecret("ACCESS_TOKEN")
		secretToken := cfg.RequireSecret("SECRET_TOKEN")

		k8accessTokenSecret, err := corev1.NewSecret(ctx, "accessToken", &corev1.SecretArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String("access-token-secret"),
			},
			StringData: pulumi.StringMap{
				"accessToken": accessToken,
			},
		})

		if err != nil {
			return err
		}

		k8secretTokenSecret, err := corev1.NewSecret(ctx, "secretToken", &corev1.SecretArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String("access-token-secret"),
			},
			StringData: pulumi.StringMap{
				"secretToken": secretToken,
			},
		})

		if err != nil {
			return err
		}

		envVars := corev1.EnvVarArray{
			corev1.EnvVarArgs{
				Name:  pulumi.String("S3_BUCKET_PATH"),
				Value: pulumi.String("ADD_S3_BUCKET_PATH_VALUE"),
			},
			corev1.EnvVarArgs{
				Name: pulumi.String("ACCESS_KEY"),
				ValueFrom: corev1.EnvVarSourceArgs{
					SecretKeyRef: corev1.SecretKeySelectorArgs{
						Name: k8accessTokenSecret.Metadata.Name(),
						Key:  pulumi.String("accessToken"),
					},
				},
			},
			corev1.EnvVarArgs{
				Name: pulumi.String("SECRET_TOKEN"),
				ValueFrom: corev1.EnvVarSourceArgs{
					SecretKeyRef: corev1.SecretKeySelectorArgs{
						Name: k8secretTokenSecret.Metadata.Name(),
						Key:  pulumi.String("secretToken"),
					},
				},
			},
		}

		commands := []string{"/usr/local/bin/content"}
		args := []string{
			"content-data",
			"--s3-bucket-path=$S3_BUCKET_PATH",
			"--access-key=$ACCESS_KEY",
			"--secret-key=$SECRET_KEY",
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
