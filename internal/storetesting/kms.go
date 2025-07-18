package storetesting

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testKms struct {
	*kms.Client

	SessionSigningKeyID                 string
	GoogleOAuthClientSecretsKMSKeyID    string
	MicrosoftOAuthClientSecretsKMSKeyID string
	GithubOAuthClientSecretsKMSKeyID    string
	OIDCClientSecretsKMSKeyID           string
	AuthenticatorAppSecretsKMSKeyID     string
}

func newKMS() (*testKms, func()) {
	container, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "nsmithuk/local-kms:latest",
				ExposedPorts: []string{"4566/tcp"},
				Env: map[string]string{
					"PORT":       "4566",
					"AWS_REGION": awsTestRegion,
				},
				Files: []testcontainers.ContainerFile{
					{
						Reader: strings.NewReader(`
# Keys generated with OpenSSL:
# openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
Keys:
  Asymmetric:
    Rsa:
      - Metadata:
          Description: "Session Signing Keys Key"
          KeyId: bc436485-5092-42b8-92a3-0aa8b93536dc
          KeyUsage: "ENCRYPT_DECRYPT"
        PrivateKeyPem: |
          -----BEGIN PRIVATE KEY-----
          MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQD21epc1564DeWZ
          80XYAXTo4tjqJzEQ6VdpkRfKHraJ4WNqS8N5HjfyzmADVOgqqlbm5M+Qq0/ViMd/
          Xqh+OUNhwvEIo6iZuNbWba3/cUV9ZFpmCv9IWvlNojc3zq0C9/fXeSqXwZWut78d
          AuFodRdAnENiHf9aXv4pIyszAxALCSCd/UCYZRw+XUDPG4pSJrwgz2Ohkqr1SnFF
          1aQt6onjt3Rtfn5IUs7BGEXGd6M3HeIlikSLjdoXEuevVaZO0ysiQdiYDYYQ2eFe
          ytXefRuotRqH4dLpL6beUFRbT1MQVtqC2S0K2wWq8T5gTFejxv6E6eVqRC2xu0lj
          TGDxnUC3AgMBAAECggEAU6K73GV69CZRS86wNbaYpGho0z4gU/ick7qD8wphE2r5
          QoUVYK6qimz+/2H/oKVC+M1Cv2Qsks/buP6b3NkOScvB3AmIET4eHV3gfRMmVoxw
          TO8g/KVGn9V9HD29Rao7ohj+I5mGXEMKUIwvUDOMg2nvMwmzAi35tHqkIo7BGtt8
          gBuuHsZj9PM6MYSSZdrHP52T3K15MaHfrLb97UaryyYnhnUmBA12DBE8MseuYA7w
          JwL3os6MwtLxRxgXnBhkk3Ist83nZNiXESXhN3d98NLS8KbX2wcbnd0B+CqRyvnv
          GbE+CfzxPf/zTsexxpS3TlTR80vAYkubmtWIMG128QKBgQD/iQbZx2xhH6VjYWC6
          +kc03povTKTe/MKUySO7poWjJrGbajrkq7RcXdNCglVSXcKY/BvmgsWRqJc+Jh2z
          enFIcGOuO146FEAr3i4hGjtV01/ukgAl6Ko68gdxjyQLqrJ/bg0qQO57KEhRh5Tb
          mR5mIkG2j2Usr4Llc3LGXIH8VQKBgQD3SNaahwum8+8kXaxgmKwfOL64rM5fLQq3
          f0UGzKZkuRSqXJn9EKuE1rNKX4zNUBWJVF+C4bjRGLz1QRS7j2taqU4awLie+5Ak
          M4Ww8lzHd3uKf+ESCd8DU3TzD+dggtuw+OTqVZdJKA5Kfrbg72ZUyzH3p9Oj/zMu
          QWl3d6TU2wKBgQCaMZs6qoWRjcEE2Ou/p+pz0qcDR6JtE+RuV3kCcJdPPbgKae2j
          sqCg49To2zCVBRK5sdc8H0kMfcjVrbZaaNYWugrMRfKz5Shb0DPRsbyAK45FrT/9
          oAmojAdF1PQRPi17i3LSPmApXMNWvxNp91lKk/1HJfwNHNNFlYZ6f7PICQKBgQCq
          q2ryXCJ+p/11a/F8+eJR6ig37YzBw6SR4RUTDEwLWHIa4q6lKsw2crhrrGbRjWRP
          1BvXiVK1fg1sd+6HRQUjHZb6f+jsUVO6qJSs+5ltUdnCTWBZwtZYxVECMQfQZICc
          NCxKT6iKpUq3v50YwiIug8+IzhwUJB5+3kacXcc14QKBgQDpjYvwAPAq1Rru/Ew4
          hzisDSCY5CLE+X/6dvogWhJBmpaZBKDmUGi6AwK9rcwITZmlR/qU+2WqNdhHxa8S
          uSp1A6OmOHQHA3I+J4veI0kPB2Y0Z65CyfCYm9MsNkcyFYx4tRBSOzAdA+xrJCa4
          y5+KYGmXlaoRhFSq1VO8mGoihA==
          -----END PRIVATE KEY-----
      - Metadata:
          Description: "Google OAuth Client Secrets Key"
          KeyId: d5670d9b-184a-400a-a08b-d05d7b99f91c
          KeyUsage: "ENCRYPT_DECRYPT"
        PrivateKeyPem: |
          -----BEGIN PRIVATE KEY-----
          MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDC0PPMEdLlX8tg
          YM4w1ExrsZMjPFfk2I136v7OK767TQbC3qFSVIJOVHRFtpNw1O5QovG9W+tAqXM4
          ZfSyUjY7DHTUEmJ5O7Bi35N2N+S8R+wXf2tEza4jEzOcvC6TGaQOyDAZaEyjyZsA
          44kihk+NmWIgljJqJooJMK5EbNM4oHCLDZStk2eF1KvwJfY9fB3eD4JAqF41dsky
          EOjMiB2wqAxmkvqc486KMFKA0ItmJrPQ7Xr+6WZDv6DU1KIylgf9h9rRnTwzfhZJ
          duBFY42PWWt31jOcYMcZPnDHO6dEDY3gE/PrpX0+WSDZVDOXRE30XLvBoCPEQJQD
          uX19n9s5AgMBAAECggEAWHubH4bA6Nk3gBC31cm24/sFPy27Jf+NUXf0PyPzPxLf
          DUccsk4b2QPWw4sHMGoly44WidDj6ryLzoPQPeXFJ9Cih2fKPhH0LRQq37jHNRTd
          kFaZG+jnPJsOCBQYe0tcDjKyVfffR0zcD+1Ibdve6gtOXEqbn0bdzwrDO+TJkp/Q
          psIPf2w8f83vFkBf7mdQVJ8X0UjzyhToB8eIWVipOWa9KxVhUzIuU1LG70jloB10
          p5VcqTIzF/CzSF+5afsQ65uPNPd7pcOmc83dEFR0jGhj/T23vzTkQxHaiMte4v0W
          dEPnf72endOAUw8uBLOjzq5TKw6ADqODk7f3ZBzquQKBgQDzDdsjQ+qGPhYATqB8
          +qztzesAc95Dvmf8jjATE5tsVNItm2l8FLAFcbQRbl54XVlHavMC3jyPG6ko9tp6
          Xv0UKzSs9UeYEpFITE0WFD7WbdLG1yVkz+v8uMbg422rF2W1qBB8ptBFGrnx1bH1
          ev/kEZ1M7Yy5+aDxiyQ1ad5NdwKBgQDNMVo58coUCiCwuIYFxJQzJH8DmB0oLrSC
          4Ig7dS3T1CNsh62yRGaePEVUZ774sPLPuWb8uVwLr83aXfW8ncWdgIN9QVT3CwjO
          RNG2zhIdG3aS9zI356385WB7YYJWqjjRaa+7+NjhR5rdbQKk3oTpkJQJWUyUSPM+
          /O4mEgiIzwKBgA4TCFSDc0owwC9mXi6+iVL/8JLHIuDDXtwmE6yXHxHn23/elv4j
          aIn4Kpgzzu6jYS8ch1PsMI+M53/Cw6YAaCFJ2zQExA+PS7BnErOrmnPqSiFPhg/P
          Znfs7z9IjCozIaWiRMojEr5drNTPLg3sAHNhfb1dqB+A0AwMpZ0eM0xDAoGBAJCA
          pgLShTYxn630dOXQ93FAzXXxhO8MXTEiAK6mqfxYlA3VSvyU8ROUbFqxqSqoKoch
          ESb/PpQ4XabfrrQDA+0UWQU3oidMDQp+KpYrb1QySAHdte7q6HuF6blaBRkVTWgk
          no6pA8s6yxQOteL3lfCKUcZ3rddrvGnqY6hJ4Nq/AoGAEsRCy/W+KtGAIv6xPT0C
          ggYwFCLglBydINRB/rLWsleAlAlalgP5hQE7RCkAfE8xEQqY6hap5qPMPnOa5ADx
          1EeZ/mbsZI0YRCLnEGjsXhcAyh/IwWUC53cPpcRDvibIwv1h2mHqLcHsBxlbYvib
          RH9NflgV5oeIi5nP+Y1DQ08=
          -----END PRIVATE KEY-----
      - Metadata:
          Description: "Microsoft OAuth Client Secrets Key"
          KeyId: e612b031-6241-4df0-8bd0-17b19995ef14
          KeyUsage: "ENCRYPT_DECRYPT"
        PrivateKeyPem: |
          -----BEGIN PRIVATE KEY-----
          MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCnLBLsD2ORW/o2
          8JhR4b6XQKCx1fif2OMOb0nxYZVT8fg2CD3oVP+XAltjszzDP9UgAXjPEsZYzZlO
          jp/WXENciLthpLWoMSeFlqNz01somEa1nDlcQ/7jLhRxYs3eh/s1iYT8mbH2Ripu
          oBs9KiNdF5actNpC++jHGG/qViVsKSQQ0HLbRUJ83xJ8XQPy4scBlYGI3BwAf7Z1
          VT8thgZDtECKyqUnSLZYVYqA1Tj9YdyRHifhVEyCZ5HbHPyCUPVfpLGub8ffBTUN
          /ax383Kud86sg43OHeUKgSCu8Oj0waDdUrAFCBdifZaz6IxyEubdZ1eS72683Ddw
          z1lzwc87AgMBAAECggEAA67w68qbAwTnynYapRM4Q9TktYZlaAA7YIILOwpPY/4c
          3fPoiUn2J7mhkdzNJfAuHfpqUwWy4RoGmriBxRNbWJqaplgeuIn8uPDMwSyTAZ35
          UN8UVHgbEZ5eTPFEX/bXDiLtjzNDvI1nOfDFKN/Yz6BJbUJ+3KL4Sgq7zIoBYRSt
          20mMkbwZSmoIlfV2bQ6u54S4ggocWHYqSrrK7SqYNSvMJFl4mLkxhYDl4UMltcsp
          1Q1tgDZ1ZL9nESydNwzgvLBt/nV0QkYmGyXa+/VFiPz8ko65Ceo/yK+k4pOxlS66
          Z/2l/rQihx32qFSaJOZa65Y3vno2I+6I0sTVNt+DoQKBgQDW5bNw3UR20zGqgBaV
          eHi/tcOVmqIeCk/o8W9IXyrG40KhIvsg3Xbj2WCu/wtf6zCfieb9vS0plaHDFSUA
          BaKm9/vJiKgydUMSSon4sYePrYQ9cCSTPjBaLWm63+lVJLkIu/icmdiUEt0dULRP
          V9YzwbCSNkdEQZ/ygsbF4vndFwKBgQDHJYuoxHG4x5TDd010aehDLs/KHJDLKBc0
          Lpvrd5JRJZWOXC2qMWQ3bMyLsz8J1cKtvk4dFfos/lwtsitUeuQ9G5MoKtdpfpsQ
          zhiN3nykMtp4fye8wPwiGXhFgOvkf9hyEFKh5MfCjqdAP+DUl6ql/9sOR76/uvBB
          3LhySbzdfQKBgQCdgSnt1R8zAEPstYjX8L5/tJcvdXDRF7nN//cSUj4mG7dgJyVs
          xyU2hsKoQGJz4Qt4Qzi8TQVm7zbqpvrBc1/thOBUrAarROrt4xgQ4P18vy6nYSRN
          j00dKx/NSgPY1duQnUTwcoocrV7G97nQVY63zITABWxiiL7UnilWLK/57QKBgQCF
          mVyOBeu86LeWQi0GEh6tI3RmxK8me2jFqxcS6o6QPcSNUq2X5bazsBuxBLkfofYO
          lQLWZG4HTUUNqt+Ct0by79LTOZp1vWfN6FV0p3O6vBrwh21jJZyAS9Hx3sFh85qD
          OwwUa+TPUuBFLBVqyazD3FdaxyrieUjBBo/+rBU2CQKBgDuVZWr/gUcvBodse1us
          85xADQ59vxgJQa9x9E8z9/2g7UC4eD0dttDSgRBaHCrp/RE1Ye5lssH5WGx276/8
          jeIcSvKQflZPrfVwQxo76a5Mn58Blz8NDVzjRYbDoZfLOkKy0Z13oSvC0NpYxgBm
          R28oFfrhJITRx5zqACZ0Cwkl
          -----END PRIVATE KEY-----
      - Metadata:
          Description: "Github OAuth Secrets Key"
          KeyId: 48f79cf6-2bbd-4f65-a934-ad1df7122fbd
          KeyUsage: "ENCRYPT_DECRYPT"
        PrivateKeyPem: |
          -----BEGIN PRIVATE KEY-----
          MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCnLBLsD2ORW/o2
          8JhR4b6XQKCx1fif2OMOb0nxYZVT8fg2CD3oVP+XAltjszzDP9UgAXjPEsZYzZlO
          jp/WXENciLthpLWoMSeFlqNz01somEa1nDlcQ/7jLhRxYs3eh/s1iYT8mbH2Ripu
          oBs9KiNdF5actNpC++jHGG/qViVsKSQQ0HLbRUJ83xJ8XQPy4scBlYGI3BwAf7Z1
          VT8thgZDtECKyqUnSLZYVYqA1Tj9YdyRHifhVEyCZ5HbHPyCUPVfpLGub8ffBTUN
          /ax383Kud86sg43OHeUKgSCu8Oj0waDdUrAFCBdifZaz6IxyEubdZ1eS72683Ddw
          z1lzwc87AgMBAAECggEAA67w68qbAwTnynYapRM4Q9TktYZlaAA7YIILOwpPY/4c
          3fPoiUn2J7mhkdzNJfAuHfpqUwWy4RoGmriBxRNbWJqaplgeuIn8uPDMwSyTAZ35
          UN8UVHgbEZ5eTPFEX/bXDiLtjzNDvI1nOfDFKN/Yz6BJbUJ+3KL4Sgq7zIoBYRSt
          20mMkbwZSmoIlfV2bQ6u54S4ggocWHYqSrrK7SqYNSvMJFl4mLkxhYDl4UMltcsp
          1Q1tgDZ1ZL9nESydNwzgvLBt/nV0QkYmGyXa+/VFiPz8ko65Ceo/yK+k4pOxlS66
          Z/2l/rQihx32qFSaJOZa65Y3vno2I+6I0sTVNt+DoQKBgQDW5bNw3UR20zGqgBaV
          eHi/tcOVmqIeCk/o8W9IXyrG40KhIvsg3Xbj2WCu/wtf6zCfieb9vS0plaHDFSUA
          BaKm9/vJiKgydUMSSon4sYePrYQ9cCSTPjBaLWm63+lVJLkIu/icmdiUEt0dULRP
          V9YzwbCSNkdEQZ/ygsbF4vndFwKBgQDHJYuoxHG4x5TDd010aehDLs/KHJDLKBc0
          Lpvrd5JRJZWOXC2qMWQ3bMyLsz8J1cKtvk4dFfos/lwtsitUeuQ9G5MoKtdpfpsQ
          zhiN3nykMtp4fye8wPwiGXhFgOvkf9hyEFKh5MfCjqdAP+DUl6ql/9sOR76/uvBB
          3LhySbzdfQKBgQCdgSnt1R8zAEPstYjX8L5/tJcvdXDRF7nN//cSUj4mG7dgJyVs
          xyU2hsKoQGJz4Qt4Qzi8TQVm7zbqpvrBc1/thOBUrAarROrt4xgQ4P18vy6nYSRN
          j00dKx/NSgPY1duQnUTwcoocrV7G97nQVY63zITABWxiiL7UnilWLK/57QKBgQCF
          mVyOBeu86LeWQi0GEh6tI3RmxK8me2jFqxcS6o6QPcSNUq2X5bazsBuxBLkfofYO
          lQLWZG4HTUUNqt+Ct0by79LTOZp1vWfN6FV0p3O6vBrwh21jJZyAS9Hx3sFh85qD
          OwwUa+TPUuBFLBVqyazD3FdaxyrieUjBBo/+rBU2CQKBgDuVZWr/gUcvBodse1us
          85xADQ59vxgJQa9x9E8z9/2g7UC4eD0dttDSgRBaHCrp/RE1Ye5lssH5WGx276/8
          jeIcSvKQflZPrfVwQxo76a5Mn58Blz8NDVzjRYbDoZfLOkKy0Z13oSvC0NpYxgBm
          R28oFfrhJITRx5zqACZ0Cwkl
          -----END PRIVATE KEY-----
      - Metadata:
          Description: "OIDC Client Secrets Key"
          KeyId: 0c261e98-324b-447b-a960-f66b8213610f
          KeyUsage: "ENCRYPT_DECRYPT"
        PrivateKeyPem: |
          -----BEGIN PRIVATE KEY-----
          MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCnLBLsD2ORW/o2
          8JhR4b6XQKCx1fif2OMOb0nxYZVT8fg2CD3oVP+XAltjszzDP9UgAXjPEsZYzZlO
          jp/WXENciLthpLWoMSeFlqNz01somEa1nDlcQ/7jLhRxYs3eh/s1iYT8mbH2Ripu
          oBs9KiNdF5actNpC++jHGG/qViVsKSQQ0HLbRUJ83xJ8XQPy4scBlYGI3BwAf7Z1
          VT8thgZDtECKyqUnSLZYVYqA1Tj9YdyRHifhVEyCZ5HbHPyCUPVfpLGub8ffBTUN
          /ax383Kud86sg43OHeUKgSCu8Oj0waDdUrAFCBdifZaz6IxyEubdZ1eS72683Ddw
          z1lzwc87AgMBAAECggEAA67w68qbAwTnynYapRM4Q9TktYZlaAA7YIILOwpPY/4c
          3fPoiUn2J7mhkdzNJfAuHfpqUwWy4RoGmriBxRNbWJqaplgeuIn8uPDMwSyTAZ35
          UN8UVHgbEZ5eTPFEX/bXDiLtjzNDvI1nOfDFKN/Yz6BJbUJ+3KL4Sgq7zIoBYRSt
          20mMkbwZSmoIlfV2bQ6u54S4ggocWHYqSrrK7SqYNSvMJFl4mLkxhYDl4UMltcsp
          1Q1tgDZ1ZL9nESydNwzgvLBt/nV0QkYmGyXa+/VFiPz8ko65Ceo/yK+k4pOxlS66
          Z/2l/rQihx32qFSaJOZa65Y3vno2I+6I0sTVNt+DoQKBgQDW5bNw3UR20zGqgBaV
          eHi/tcOVmqIeCk/o8W9IXyrG40KhIvsg3Xbj2WCu/wtf6zCfieb9vS0plaHDFSUA
          BaKm9/vJiKgydUMSSon4sYePrYQ9cCSTPjBaLWm63+lVJLkIu/icmdiUEt0dULRP
          V9YzwbCSNkdEQZ/ygsbF4vndFwKBgQDHJYuoxHG4x5TDd010aehDLs/KHJDLKBc0
          Lpvrd5JRJZWOXC2qMWQ3bMyLsz8J1cKtvk4dFfos/lwtsitUeuQ9G5MoKtdpfpsQ
          zhiN3nykMtp4fye8wPwiGXhFgOvkf9hyEFKh5MfCjqdAP+DUl6ql/9sOR76/uvBB
          3LhySbzdfQKBgQCdgSnt1R8zAEPstYjX8L5/tJcvdXDRF7nN//cSUj4mG7dgJyVs
          xyU2hsKoQGJz4Qt4Qzi8TQVm7zbqpvrBc1/thOBUrAarROrt4xgQ4P18vy6nYSRN
          j00dKx/NSgPY1duQnUTwcoocrV7G97nQVY63zITABWxiiL7UnilWLK/57QKBgQCF
          mVyOBeu86LeWQi0GEh6tI3RmxK8me2jFqxcS6o6QPcSNUq2X5bazsBuxBLkfofYO
          lQLWZG4HTUUNqt+Ct0by79LTOZp1vWfN6FV0p3O6vBrwh21jJZyAS9Hx3sFh85qD
          OwwUa+TPUuBFLBVqyazD3FdaxyrieUjBBo/+rBU2CQKBgDuVZWr/gUcvBodse1us
          85xADQ59vxgJQa9x9E8z9/2g7UC4eD0dttDSgRBaHCrp/RE1Ye5lssH5WGx276/8
          jeIcSvKQflZPrfVwQxo76a5Mn58Blz8NDVzjRYbDoZfLOkKy0Z13oSvC0NpYxgBm
          R28oFfrhJITRx5zqACZ0Cwkl
          -----END PRIVATE KEY-----
      - Metadata:
          Description: "Authenticator Apps Secrets Key"
          KeyId: 186bf523-bc6a-479b-8ed1-e198298b98f2
          KeyUsage: "ENCRYPT_DECRYPT"
        PrivateKeyPem: |
          -----BEGIN PRIVATE KEY-----
          MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCnLBLsD2ORW/o2
          8JhR4b6XQKCx1fif2OMOb0nxYZVT8fg2CD3oVP+XAltjszzDP9UgAXjPEsZYzZlO
          jp/WXENciLthpLWoMSeFlqNz01somEa1nDlcQ/7jLhRxYs3eh/s1iYT8mbH2Ripu
          oBs9KiNdF5actNpC++jHGG/qViVsKSQQ0HLbRUJ83xJ8XQPy4scBlYGI3BwAf7Z1
          VT8thgZDtECKyqUnSLZYVYqA1Tj9YdyRHifhVEyCZ5HbHPyCUPVfpLGub8ffBTUN
          /ax383Kud86sg43OHeUKgSCu8Oj0waDdUrAFCBdifZaz6IxyEubdZ1eS72683Ddw
          z1lzwc87AgMBAAECggEAA67w68qbAwTnynYapRM4Q9TktYZlaAA7YIILOwpPY/4c
          3fPoiUn2J7mhkdzNJfAuHfpqUwWy4RoGmriBxRNbWJqaplgeuIn8uPDMwSyTAZ35
          UN8UVHgbEZ5eTPFEX/bXDiLtjzNDvI1nOfDFKN/Yz6BJbUJ+3KL4Sgq7zIoBYRSt
          20mMkbwZSmoIlfV2bQ6u54S4ggocWHYqSrrK7SqYNSvMJFl4mLkxhYDl4UMltcsp
          1Q1tgDZ1ZL9nESydNwzgvLBt/nV0QkYmGyXa+/VFiPz8ko65Ceo/yK+k4pOxlS66
          Z/2l/rQihx32qFSaJOZa65Y3vno2I+6I0sTVNt+DoQKBgQDW5bNw3UR20zGqgBaV
          eHi/tcOVmqIeCk/o8W9IXyrG40KhIvsg3Xbj2WCu/wtf6zCfieb9vS0plaHDFSUA
          BaKm9/vJiKgydUMSSon4sYePrYQ9cCSTPjBaLWm63+lVJLkIu/icmdiUEt0dULRP
          V9YzwbCSNkdEQZ/ygsbF4vndFwKBgQDHJYuoxHG4x5TDd010aehDLs/KHJDLKBc0
          Lpvrd5JRJZWOXC2qMWQ3bMyLsz8J1cKtvk4dFfos/lwtsitUeuQ9G5MoKtdpfpsQ
          zhiN3nykMtp4fye8wPwiGXhFgOvkf9hyEFKh5MfCjqdAP+DUl6ql/9sOR76/uvBB
          3LhySbzdfQKBgQCdgSnt1R8zAEPstYjX8L5/tJcvdXDRF7nN//cSUj4mG7dgJyVs
          xyU2hsKoQGJz4Qt4Qzi8TQVm7zbqpvrBc1/thOBUrAarROrt4xgQ4P18vy6nYSRN
          j00dKx/NSgPY1duQnUTwcoocrV7G97nQVY63zITABWxiiL7UnilWLK/57QKBgQCF
          mVyOBeu86LeWQi0GEh6tI3RmxK8me2jFqxcS6o6QPcSNUq2X5bazsBuxBLkfofYO
          lQLWZG4HTUUNqt+Ct0by79LTOZp1vWfN6FV0p3O6vBrwh21jJZyAS9Hx3sFh85qD
          OwwUa+TPUuBFLBVqyazD3FdaxyrieUjBBo/+rBU2CQKBgDuVZWr/gUcvBodse1us
          85xADQ59vxgJQa9x9E8z9/2g7UC4eD0dttDSgRBaHCrp/RE1Ye5lssH5WGx276/8
          jeIcSvKQflZPrfVwQxo76a5Mn58Blz8NDVzjRYbDoZfLOkKy0Z13oSvC0NpYxgBm
          R28oFfrhJITRx5zqACZ0Cwkl
          -----END PRIVATE KEY-----
Aliases:
  - AliasName: alias/session-signing-keys-key
    TargetKeyId: bc436485-5092-42b8-92a3-0aa8b93536dc
  - AliasName: alias/google-oauth-client-secrets-key
    TargetKeyId: d5670d9b-184a-400a-a08b-d05d7b99f91c
  - AliasName: alias/microsoft-oauth-client-secrets-key
    TargetKeyId: e612b031-6241-4df0-8bd0-17b19995ef14
  - AliasName: alias/github-oauth-client-secrets-key
    TargetKeyId: 48f79cf6-2bbd-4f65-a934-ad1df7122fbd
  - AliasName: alias/oidc-client-secrets-key
    TargetKeyId: 0c261e98-324b-447b-a960-f66b8213610f
  - AliasName: alias/authenticator-apps-secrets-key
    TargetKeyId: 186bf523-bc6a-479b-8ed1-e198298b98f2
`),
						ContainerFilePath: "/init/seed.yaml",
					},
				},
				WaitingFor: wait.ForLog("Local KMS started on 0.0.0.0:4566"),
			},
			Started: true,
		},
	)
	cleanup := func() {
		_ = testcontainers.TerminateContainer(container)
	}
	if err != nil {
		cleanup()
		log.Panicf("run local KMS container: %v", err)
	}
	endpoint, err := container.PortEndpoint(context.Background(), "4566/tcp", "http")
	if err != nil {
		cleanup()
		log.Panicf("get local KMS endpoint: %v", err)
	}
	cfg := kms.Options{
		Region:       awsTestRegion,
		BaseEndpoint: &endpoint,
	}
	return &testKms{
		Client:                              kms.New(cfg),
		SessionSigningKeyID:                 "bc436485-5092-42b8-92a3-0aa8b93536dc",
		GoogleOAuthClientSecretsKMSKeyID:    "d5670d9b-184a-400a-a08b-d05d7b99f91c",
		MicrosoftOAuthClientSecretsKMSKeyID: "e612b031-6241-4df0-8bd0-17b19995ef14",
		GithubOAuthClientSecretsKMSKeyID:    "48f79cf6-2bbd-4f65-a934-ad1df7122fbd",
		OIDCClientSecretsKMSKeyID:           "0c261e98-324b-447b-a960-f66b8213610f",
		AuthenticatorAppSecretsKMSKeyID:     "186bf523-bc6a-479b-8ed1-e198298b98f2",
	}, cleanup
}
