package main

import (
	"fmt"

	batchv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/batch/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		jobName := "modware-import"
		jobLabels := pulumi.StringMap{
			"app": pulumi.String(jobName),
		}
		cfg := config.New(ctx, "")
		accessToken := cfg.RequireSecret("ACCESS_TOKEN")
		secretToken := cfg.RequireSecret("SECRET_TOKEN")
		imageTag := cfg.Require("IMAGE_TAG")
		s3BucketPath := cfg.Require("S3_BUCKET_PATH")

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
				Value: pulumi.String(s3BucketPath),
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

		job, err := batchv1.NewJob(ctx, jobName, &batchv1.JobArgs{
			Spec: batchv1.JobSpecArgs{
				Template: &corev1.PodTemplateSpecArgs{
					Metadata: &metav1.ObjectMetaArgs{
						Labels: jobLabels,
					},
					Spec: &corev1.PodSpecArgs{
						RestartPolicy: pulumi.String("Never"),
						Containers: corev1.ContainerArray{
							corev1.ContainerArgs{
								Name:    pulumi.String(jobName),
								Image:   pulumi.String(fmt.Sprintf("dictybase/modware-import:%s", imageTag)),
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

		ctx.Export("name", job.Metadata.Name())
		return nil
	})
}
