/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package kubeconfig

import (
	"bytes"
	util "eke/internal/util/utilityFunctions"
	b64 "encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// template for kubectl config file, used in dynamic authentication:
// i.e. automatic renewal of cert and key after expiration is done
// where kubectl uses the "eke kubeconfig auth" command behind the scene
var (
	kubectlConfigTemplate = template.Must(template.New("kubectl-config").Parse(`apiVersion: v1
kind: Config
users:
- name: {{.Signum}}
  user:
    exec:
      command: "eke"
      apiVersion: "client.authentication.k8s.io/v1beta1"
      args:
      - "kubeconfig"
      - "auth"
clusters:
- name: {{.ClusterName}}
  cluster:
    server: "{{.APIserverEndpoint}}"
    certificate-authority-data:  LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZ5RENDQTdDZ0F3SUJBZ0lSQU9uSXdMZ1V0dnhCaDgyTTFnMVdMRWN3RFFZSktvWklodmNOQVFFTEJRQXcKYlRFTE1Ba0dBMVVFQmhNQ1UwVXhFakFRQmdOVkJBZ01DVk4wYjJOcmFHOXNiVEVTTUJBR0ExVUVCd3dKVTNSdgpZMnRvYjJ4dE1SRXdEd1lEVlFRS0RBaEZjbWxqYzNOdmJqRU5NQXNHQTFVRUN3d0VRMDVFUlRFVU1CSUdBMVVFCkF3d0xSVmRUSUZKdmIzUWdRMEV3SGhjTk1qQXdOVEUzTWpBek9UQTRXaGNOTkRVd05URXhNakF6T1RBNFdqQnQKTVFzd0NRWURWUVFHRXdKVFJURVNNQkFHQTFVRUNBd0pVM1J2WTJ0b2IyeHRNUkl3RUFZRFZRUUhEQWxUZEc5agphMmh2YkcweEVUQVBCZ05WQkFvTUNFVnlhV056YzI5dU1RMHdDd1lEVlFRTERBUkRUa1JGTVJRd0VnWURWUVFECkRBdEZWMU1nVW05dmRDQkRRVENDQWlJd0RRWUpLb1pJaHZjTkFRRUJCUUFEZ2dJUEFEQ0NBZ29DZ2dJQkFNQ2kKbDluczczYW9Cam9oRzhaSDlZeVhWNWQ4UUw5ZmlGRC96MThCcU9ZTGZtVFBlM01zMHJrcmdkUUJIMUdib1hNQwpJbzJoVFFERi9sMzJXWFlWcXpPU3BvUzhNdkR2MFNaRGFUNW1QeWdQZVozSU5ndmZwNldnZnV2VE9MUG1sWEY5CnRCaXdQSU9iMGh3RkxtOVQrTW5ISW5mbG0wZGJxYXhxT2ZsQ3ltbDBkSCtiQ1l3WmxKa1VXUXI1SThyUUxtN1MKSzBneXFYMHE1VTR5NTF6TnpZRlZmWWZFSGNTbDBnZEN3ekhOaDc0ekl4aktQRmxBbVNNVTRES1hURXBqdDQzTgpZR0tYUk9DWUtkZmlzVWRGQlhVTnhkNzVGMXNwWEVBblZUMlVVWk1ZelhidzRxR2NzUDhpUFVrUXBJUEU5aG16CllJMUpJOUVvaWJnajNOV0hxMGdzRTJIdUdOdDFoeVhpVmhGTytuYkw0L21iY250OUxuT2sra0txZXNlZVNSd08KLzVqTFJITTY4MVptdUwrTXFaSm4zR2FyY2xqNzRWMFVrbFc5ZUU0a3NTeVEzQ1hrN24vWWlBUXpTUWZkV0wwWgpWTHpiTkxEdC93QzVPRXBuUG5DbnhYUXJzVzlQRnlGeDNoWkV1a3FFMENzY0drK1NCL1R6YXNrb1pvTU0xV3JDCndITXEzNUxnN1BJWW9UNTZXaHFJeHFRSzBrNitBbkJORUc0d043ODdVOUlsbW1WdHlzdjFDbkVrMW5oU3A0MjgKSjBnWUZKZDltNUdGU3BIYk44NDE2OVRiZjlXNVBkMVV4andVc2JHMlVvOWkyQWZZYkxrYVZ4OVUvcHhtZmpkVwpYVnpPbjZCQ0liUlFjaVpSeVp2czE0ak1vaElOclBIYVRYOTM0VnJWQWdNQkFBR2pZekJoTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RnWURWUjBQQVFIL0JBUURBZ0dHTUIwR0ExVWREZ1FXQkJSYmRqU1JLdlJxcW12NG8zM3kKVDk3ay9TQkkvREFmQmdOVkhTTUVHREFXZ0JSYmRqU1JLdlJxcW12NG8zM3lUOTdrL1NCSS9EQU5CZ2txaGtpRwo5dzBCQVFzRkFBT0NBZ0VBb2RHMzFjQUxQRzVOUkZkSVZMc2hOK2EyQzcyQVk4WnNVSEx6OERHOEpqb3ZVQ1VwCjlRL3NOSnB3eVY4WGtiTG91Wjh2WUdFRms3RUFzYWRIdkRDa0dqZ0lPVFI4NnlGdlJxbFkraVZ6Q2xYd0xpMlQKYWFodTQ4QnV0bVhqQlU4WjIxQkNySUF4aTg4Z01aUVQ3dkl0eHN4WG1iU1NmZHhFdDh1ek9ESUxEV0twU2lVagpEUzZmbmNCN1psNUlGWk9tbVhSaERieHEwbFl3RlZxOEQ5RXQ3QTM4UmhHUzQ3SVJXRTZDeFBNdldvSkREYjRKCkxKcmdVU0JEZWUrY0VwMUtQS1BwcUZpVjV1TE5hV2JJK1NaQkZnLzkwbTk5WlAxZWIxd0dhMmE2NjZCN2xnTVAKT2Z1S1llT1IySU9UclptQWJiTXMrOEthV2lHNHFlakFzaHROMDRQcDdEN2djdXJ5VTJlSGZKS0ZHeGhsMkNzbQpYQXFRdmMzM1NtN2xiVDN6VFlZUEphWXR6N1ZZMXNNL29vc1ozdk9JTGloVDZvYnJZeENRWElHeUpIMEFvSDg2CmJrSzNhSE9aWW5qS0dhaEZXb2xmcFdKeXpSaVZ1KytwQVpaVXU2VjJQM2RUakI3TVlOTG1LQmFlZnJRbVhNT1YKdEUxQnVFKy9yalNSNzhuTEc4a3dyVk1xZkxyRHRsK1JxcE9ET0oxdnpONHRxKzM2dTNHMnRJVEdoVTh1SlJPMApieS9QVGxMaFc5TUd4SWwzSTk3a2xZT2dMSXpVdWh4a3ZCZXgvUDVaMk1NaHJxS0N2TW9RZ1ZVclFjWHlYcGF6CjMyVVNKK1dYTzc5Ty83WVExM2lpTnlXQWZQQlBtWXpRR2k3OGVlR1dtWkc2L01ReHliTjJkL1AyMXFBPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tQkVHSU4gQ0VSVElGSUNBVEUtLS0tLQpNSUlFeWpDQ0FyS2dBd0lCQWdJUkFPbkl3TGdVdHZ4Qmg4Mk0xZzFXTEVnd0RRWUpLb1pJaHZjTkFRRUxCUUF3CmJURUxNQWtHQTFVRUJoTUNVMFV4RWpBUUJnTlZCQWdNQ1ZOMGIyTnJhRzlzYlRFU01CQUdBMVVFQnd3SlUzUnYKWTJ0b2IyeHRNUkV3RHdZRFZRUUtEQWhGY21samMzTnZiakVOTUFzR0ExVUVDd3dFUTA1RVJURVVNQklHQTFVRQpBd3dMUlZkVElGSnZiM1FnUTBFd0hoY05NakF3TlRFNE1ETTFNakl6V2hjTk16QXdOVEUyTURNMU1qSXpXakJzCk1Rc3dDUVlEVlFRR0V3SlRSVEVTTUJBR0ExVUVDQXdKVTNSdlkydG9iMnh0TVJJd0VBWURWUVFIREFsVGRHOWoKYTJodmJHMHhFVEFQQmdOVkJBb01DRVZ5YVdOemMyOXVNUTB3Q3dZRFZRUUxEQVJEVGtSRk1STXdFUVlEVlFRRApEQXByZFdKbGNtNWxkR1Z6TUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQ0FRRUF2NWdhClhaR2hvVSs5MXFNUmp6aXdGWTNFeUZXZEpzbE1Yb1ZWbHdKSEkza09EWFZVcTJUNEhSd3N4cnUzeEVtVkcvQkcKMWpWcGFPR0JlWUMzNEhjeERvNVh3RURIazMvV0QyYmtxTllwUEswU3BBMXk1cUEzS2FGOUxCM05rRUdQU0tCZgphSUVOUVptV3pNN2RuZUtPM1p0V3RGUEZkeWF4WEp2d0kwaHR2cy81eExCM2tMWHZOdkRrRjVRQ0NIdXAxUU5kCmlBb3dmMnlQWmsyN3pmN1JiYysxUDlvZjcwRG5RY2k0T1A1K2g5Qis0cEJNamJ3MVRDQ1N0SE9LUzBEOForL0UKbFR0bUN0ajVTRVhodFJnd2JXQVhaNStyb1lhTnBjRjRVbVJ1TmNSWFlVSzhGOUtnOGFYYi94all6cUQ1c3dRRgp2TFAwZ1hmRUJMdVgycC9GU1FJREFRQUJvMll3WkRBU0JnTlZIUk1CQWY4RUNEQUdBUUgvQWdFQU1BNEdBMVVkCkR3RUIvd1FFQXdJQmhqQWRCZ05WSFE0RUZnUVVBUFgycHV1SXpvT012THhlNWJud1djaUE4dFF3SHdZRFZSMGoKQkJnd0ZvQVVXM1kwa1NyMGFxcHIrS045OGsvZTVQMGdTUHd3RFFZSktvWklodmNOQVFFTEJRQURnZ0lCQUF0cgpHbW1zcldvMFRmNUtPOU9JcEo0dlNLdEZNTUR6VWEvOEVIZHg4WG1TbXZxS2YrajE2TXg0cUFwaUE2ZHR4ODlJCjdSMlkrd2pKWWlMOGV0c2tQVGlLdGNuV0JDNEJzNUpZNTFsblhjenFnbG91cE5hTnNRV0FTOEZySldwU0xMU0kKTkU0anhEY1ZyajBuaG1KeEVUZ3FkSmRPVTByK2FtMXFmeHNKQ0dNa0tMVTgxaE12UHRnWFUrK01oVjVwOXhaQgpLdklLVHlHbnlGWnpUZ3BEWXdxU3doSTRhRmNieG1qcGkwMUtiaHFXZWJNSjVzOFVZZFFZSm4weWxSd2NzMTB5Cjd2MEQvcEhmREVRSzFFVnNhd0haTlZVaW9kRWw1VUFuTzBudWcwWDlzeFJmeGNQRmtOSmMwUk1UMFVJRFlWMnAKMnluM2pCZkZlWVhDazdiZURzVmpyZEw5NkR2aHd3V3UxNmpTZWlNc2lWdUpHYzZWNHRkZEF0VUVxenJFOGZwaApHS2F2eXh2dVJsemFPMnJURkw2ZUdSWkhtaWwrQ04vQ0sraWNRQWF4WWZmWnl6YVUwbUVydEtpUnlDOW93RUFOClAzc0trSS93RGJibmxkeHE5Wnp6Q0plclNFcFdHUi9HVExqR0t4cFV2NGRSR1RtWXVOZjU1ZVU4Zzk5YXdrdnMKTHc3SlFZeng3bVpacnVxclNQS3QrVHZCSkp2M0hMbVlTY09FTlk3VUJvMTliWTl1bFVnSHpGaDFtUVRDaCtmNgpiSUY3RUF5eGtnb2ZVeXVpbWQwVTJHaTZjd2ZPbGdFUFkzWW5oTVdxUExFMlhmbG5Hbk9pN1h1VFFDWFl2bGF1CktUQ28vZVJhcmQ1ZGNlSzZkb3A4NklhRFNpWEZiSVZlTnRPcHI1bnkKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
contexts:
- name: {{.ClusterName}}
  context:
    cluster: {{.ClusterName}}
    user: {{.Signum}}
current-context: {{.ClusterName}}
`))
	// template for static kubeconfig file, used in static authentication
	// i.e. manual renewal of cert and key after expiration is needed
	staticConfigTemplate = template.Must(template.New("static-config").Parse(`apiVersion: v1
kind: Config
users:
- name: {{.Signum}}
  user:
    client-certificate-data: {{.ClientCert}}
    client-key-data: {{.ClientKey}}
clusters:
- name: {{.ClusterName}}
  cluster:
    server: "{{.APIserverEndpoint}}"
    certificate-authority-data:  LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZ5RENDQTdDZ0F3SUJBZ0lSQU9uSXdMZ1V0dnhCaDgyTTFnMVdMRWN3RFFZSktvWklodmNOQVFFTEJRQXcKYlRFTE1Ba0dBMVVFQmhNQ1UwVXhFakFRQmdOVkJBZ01DVk4wYjJOcmFHOXNiVEVTTUJBR0ExVUVCd3dKVTNSdgpZMnRvYjJ4dE1SRXdEd1lEVlFRS0RBaEZjbWxqYzNOdmJqRU5NQXNHQTFVRUN3d0VRMDVFUlRFVU1CSUdBMVVFCkF3d0xSVmRUSUZKdmIzUWdRMEV3SGhjTk1qQXdOVEUzTWpBek9UQTRXaGNOTkRVd05URXhNakF6T1RBNFdqQnQKTVFzd0NRWURWUVFHRXdKVFJURVNNQkFHQTFVRUNBd0pVM1J2WTJ0b2IyeHRNUkl3RUFZRFZRUUhEQWxUZEc5agphMmh2YkcweEVUQVBCZ05WQkFvTUNFVnlhV056YzI5dU1RMHdDd1lEVlFRTERBUkRUa1JGTVJRd0VnWURWUVFECkRBdEZWMU1nVW05dmRDQkRRVENDQWlJd0RRWUpLb1pJaHZjTkFRRUJCUUFEZ2dJUEFEQ0NBZ29DZ2dJQkFNQ2kKbDluczczYW9Cam9oRzhaSDlZeVhWNWQ4UUw5ZmlGRC96MThCcU9ZTGZtVFBlM01zMHJrcmdkUUJIMUdib1hNQwpJbzJoVFFERi9sMzJXWFlWcXpPU3BvUzhNdkR2MFNaRGFUNW1QeWdQZVozSU5ndmZwNldnZnV2VE9MUG1sWEY5CnRCaXdQSU9iMGh3RkxtOVQrTW5ISW5mbG0wZGJxYXhxT2ZsQ3ltbDBkSCtiQ1l3WmxKa1VXUXI1SThyUUxtN1MKSzBneXFYMHE1VTR5NTF6TnpZRlZmWWZFSGNTbDBnZEN3ekhOaDc0ekl4aktQRmxBbVNNVTRES1hURXBqdDQzTgpZR0tYUk9DWUtkZmlzVWRGQlhVTnhkNzVGMXNwWEVBblZUMlVVWk1ZelhidzRxR2NzUDhpUFVrUXBJUEU5aG16CllJMUpJOUVvaWJnajNOV0hxMGdzRTJIdUdOdDFoeVhpVmhGTytuYkw0L21iY250OUxuT2sra0txZXNlZVNSd08KLzVqTFJITTY4MVptdUwrTXFaSm4zR2FyY2xqNzRWMFVrbFc5ZUU0a3NTeVEzQ1hrN24vWWlBUXpTUWZkV0wwWgpWTHpiTkxEdC93QzVPRXBuUG5DbnhYUXJzVzlQRnlGeDNoWkV1a3FFMENzY0drK1NCL1R6YXNrb1pvTU0xV3JDCndITXEzNUxnN1BJWW9UNTZXaHFJeHFRSzBrNitBbkJORUc0d043ODdVOUlsbW1WdHlzdjFDbkVrMW5oU3A0MjgKSjBnWUZKZDltNUdGU3BIYk44NDE2OVRiZjlXNVBkMVV4andVc2JHMlVvOWkyQWZZYkxrYVZ4OVUvcHhtZmpkVwpYVnpPbjZCQ0liUlFjaVpSeVp2czE0ak1vaElOclBIYVRYOTM0VnJWQWdNQkFBR2pZekJoTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RnWURWUjBQQVFIL0JBUURBZ0dHTUIwR0ExVWREZ1FXQkJSYmRqU1JLdlJxcW12NG8zM3kKVDk3ay9TQkkvREFmQmdOVkhTTUVHREFXZ0JSYmRqU1JLdlJxcW12NG8zM3lUOTdrL1NCSS9EQU5CZ2txaGtpRwo5dzBCQVFzRkFBT0NBZ0VBb2RHMzFjQUxQRzVOUkZkSVZMc2hOK2EyQzcyQVk4WnNVSEx6OERHOEpqb3ZVQ1VwCjlRL3NOSnB3eVY4WGtiTG91Wjh2WUdFRms3RUFzYWRIdkRDa0dqZ0lPVFI4NnlGdlJxbFkraVZ6Q2xYd0xpMlQKYWFodTQ4QnV0bVhqQlU4WjIxQkNySUF4aTg4Z01aUVQ3dkl0eHN4WG1iU1NmZHhFdDh1ek9ESUxEV0twU2lVagpEUzZmbmNCN1psNUlGWk9tbVhSaERieHEwbFl3RlZxOEQ5RXQ3QTM4UmhHUzQ3SVJXRTZDeFBNdldvSkREYjRKCkxKcmdVU0JEZWUrY0VwMUtQS1BwcUZpVjV1TE5hV2JJK1NaQkZnLzkwbTk5WlAxZWIxd0dhMmE2NjZCN2xnTVAKT2Z1S1llT1IySU9UclptQWJiTXMrOEthV2lHNHFlakFzaHROMDRQcDdEN2djdXJ5VTJlSGZKS0ZHeGhsMkNzbQpYQXFRdmMzM1NtN2xiVDN6VFlZUEphWXR6N1ZZMXNNL29vc1ozdk9JTGloVDZvYnJZeENRWElHeUpIMEFvSDg2CmJrSzNhSE9aWW5qS0dhaEZXb2xmcFdKeXpSaVZ1KytwQVpaVXU2VjJQM2RUakI3TVlOTG1LQmFlZnJRbVhNT1YKdEUxQnVFKy9yalNSNzhuTEc4a3dyVk1xZkxyRHRsK1JxcE9ET0oxdnpONHRxKzM2dTNHMnRJVEdoVTh1SlJPMApieS9QVGxMaFc5TUd4SWwzSTk3a2xZT2dMSXpVdWh4a3ZCZXgvUDVaMk1NaHJxS0N2TW9RZ1ZVclFjWHlYcGF6CjMyVVNKK1dYTzc5Ty83WVExM2lpTnlXQWZQQlBtWXpRR2k3OGVlR1dtWkc2L01ReHliTjJkL1AyMXFBPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tQkVHSU4gQ0VSVElGSUNBVEUtLS0tLQpNSUlFeWpDQ0FyS2dBd0lCQWdJUkFPbkl3TGdVdHZ4Qmg4Mk0xZzFXTEVnd0RRWUpLb1pJaHZjTkFRRUxCUUF3CmJURUxNQWtHQTFVRUJoTUNVMFV4RWpBUUJnTlZCQWdNQ1ZOMGIyTnJhRzlzYlRFU01CQUdBMVVFQnd3SlUzUnYKWTJ0b2IyeHRNUkV3RHdZRFZRUUtEQWhGY21samMzTnZiakVOTUFzR0ExVUVDd3dFUTA1RVJURVVNQklHQTFVRQpBd3dMUlZkVElGSnZiM1FnUTBFd0hoY05NakF3TlRFNE1ETTFNakl6V2hjTk16QXdOVEUyTURNMU1qSXpXakJzCk1Rc3dDUVlEVlFRR0V3SlRSVEVTTUJBR0ExVUVDQXdKVTNSdlkydG9iMnh0TVJJd0VBWURWUVFIREFsVGRHOWoKYTJodmJHMHhFVEFQQmdOVkJBb01DRVZ5YVdOemMyOXVNUTB3Q3dZRFZRUUxEQVJEVGtSRk1STXdFUVlEVlFRRApEQXByZFdKbGNtNWxkR1Z6TUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQ0FRRUF2NWdhClhaR2hvVSs5MXFNUmp6aXdGWTNFeUZXZEpzbE1Yb1ZWbHdKSEkza09EWFZVcTJUNEhSd3N4cnUzeEVtVkcvQkcKMWpWcGFPR0JlWUMzNEhjeERvNVh3RURIazMvV0QyYmtxTllwUEswU3BBMXk1cUEzS2FGOUxCM05rRUdQU0tCZgphSUVOUVptV3pNN2RuZUtPM1p0V3RGUEZkeWF4WEp2d0kwaHR2cy81eExCM2tMWHZOdkRrRjVRQ0NIdXAxUU5kCmlBb3dmMnlQWmsyN3pmN1JiYysxUDlvZjcwRG5RY2k0T1A1K2g5Qis0cEJNamJ3MVRDQ1N0SE9LUzBEOForL0UKbFR0bUN0ajVTRVhodFJnd2JXQVhaNStyb1lhTnBjRjRVbVJ1TmNSWFlVSzhGOUtnOGFYYi94all6cUQ1c3dRRgp2TFAwZ1hmRUJMdVgycC9GU1FJREFRQUJvMll3WkRBU0JnTlZIUk1CQWY4RUNEQUdBUUgvQWdFQU1BNEdBMVVkCkR3RUIvd1FFQXdJQmhqQWRCZ05WSFE0RUZnUVVBUFgycHV1SXpvT012THhlNWJud1djaUE4dFF3SHdZRFZSMGoKQkJnd0ZvQVVXM1kwa1NyMGFxcHIrS045OGsvZTVQMGdTUHd3RFFZSktvWklodmNOQVFFTEJRQURnZ0lCQUF0cgpHbW1zcldvMFRmNUtPOU9JcEo0dlNLdEZNTUR6VWEvOEVIZHg4WG1TbXZxS2YrajE2TXg0cUFwaUE2ZHR4ODlJCjdSMlkrd2pKWWlMOGV0c2tQVGlLdGNuV0JDNEJzNUpZNTFsblhjenFnbG91cE5hTnNRV0FTOEZySldwU0xMU0kKTkU0anhEY1ZyajBuaG1KeEVUZ3FkSmRPVTByK2FtMXFmeHNKQ0dNa0tMVTgxaE12UHRnWFUrK01oVjVwOXhaQgpLdklLVHlHbnlGWnpUZ3BEWXdxU3doSTRhRmNieG1qcGkwMUtiaHFXZWJNSjVzOFVZZFFZSm4weWxSd2NzMTB5Cjd2MEQvcEhmREVRSzFFVnNhd0haTlZVaW9kRWw1VUFuTzBudWcwWDlzeFJmeGNQRmtOSmMwUk1UMFVJRFlWMnAKMnluM2pCZkZlWVhDazdiZURzVmpyZEw5NkR2aHd3V3UxNmpTZWlNc2lWdUpHYzZWNHRkZEF0VUVxenJFOGZwaApHS2F2eXh2dVJsemFPMnJURkw2ZUdSWkhtaWwrQ04vQ0sraWNRQWF4WWZmWnl6YVUwbUVydEtpUnlDOW93RUFOClAzc0trSS93RGJibmxkeHE5Wnp6Q0plclNFcFdHUi9HVExqR0t4cFV2NGRSR1RtWXVOZjU1ZVU4Zzk5YXdrdnMKTHc3SlFZeng3bVpacnVxclNQS3QrVHZCSkp2M0hMbVlTY09FTlk3VUJvMTliWTl1bFVnSHpGaDFtUVRDaCtmNgpiSUY3RUF5eGtnb2ZVeXVpbWQwVTJHaTZjd2ZPbGdFUFkzWW5oTVdxUExFMlhmbG5Hbk9pN1h1VFFDWFl2bGF1CktUQ28vZVJhcmQ1ZGNlSzZkb3A4NklhRFNpWEZiSVZlTnRPcHI1bnkKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
contexts:
- name: {{.ClusterName}}
  context:
    cluster: {{.ClusterName}}
    user: {{.Signum}}
current-context: {{.ClusterName}}
`))

	// for signum and password flags
	signum, pass string
)

// This command also prints out user roles using kubectl in the end
func kubeconfigInitCmd() *cobra.Command {
	// initCmd represents the init command
	var initCmd = &cobra.Command{
		Use:   "init <cluster name>",
		Short: "Initialize the kubeconfig file for kubectl",
		Long: `Call the EWS to get the API server endpoint of a cluster
		and then creates the kubeconfig file for kubectl.`,
		Run: func(cmd *cobra.Command, args []string) {

			// Check for the name of the cluster
			var err error
			var clusterName string

			if len(args) != 1 {
				log.Fatal("please enter the cluster name: eke kubeconfig init <cluster name>")
			} else {
				clusterName = args[0]
			}

			// We can give cusotmized name for the kubeconfig file.
			// The name of the kubeconfig file is set via the --kubeconfig flag,
			// or KUBECONFIG env variable, or the default path(aka config)
			// 1. Read the flag
			var kubeconfig_path string
			kubeconfig_path, _ = cmd.Flags().GetString("kubeconfig")
			if kubeconfig_path == "" {
				// 2. Read the env variable
				kubeconfig_path = os.Getenv("KUBECONFIG")
				if kubeconfig_path == "" { // 3. The last option is to use the default path
					kubeconfig_path = util.Get_kubeconfig_path() + "config"
				}
			}

			// Get signum and password from the user if not already set by corresponding flags
			if signum == "" {
				signum, err = util.GetUserSignum()
				if err != nil {
					log.Fatalln("error occured while prompting for credentials:", err)
				}
			}

			if pass == "" {
				pass, err = util.GetUserPassword()
				if err != nil {
					log.Fatalln("error occured while prompting for credentials:", err)
				}
			}

			// Cache user .crt and .key file into the given location
			eke_cache := util.Get_eke_path()
			util.RequestCertAndKeyFromEWS(eke_cache, signum, pass, false)

			// Check if static kubeconfig file requested
			var staticConfig bool
			staticConfig, _ = cmd.Flags().GetBool("static")
			if staticConfig {
				userCert, userKey := util.GetCertAndKey(eke_cache)
				createStaticConfig(getAPIserverEndpoint(clusterName), clusterName, signum, kubeconfig_path, userCert, userKey)
			} else {
				// Create and save a config file for kubectl to dynamically take care of user authentication
				createKubectlConfig(getAPIserverEndpoint(clusterName), clusterName, signum, kubeconfig_path)
				_, err = exec.LookPath("eke")
				if err != nil {
					log.Fatalln(err, ". before use, please add eke client in your user PATH!")
				}
			}

			// ********* Inform user of their roles *************
			// first check if kubectl command exists
			_, err = exec.LookPath("kubectl")
			if err != nil {
				log.Fatalln("not able to fetch user access:", err, ". please contact cluster owner if you don't have access or any is missing.")
			}
			// execute kubectl and print out the user role information
			kubectlCmd := exec.Command("kubectl", "--kubeconfig", kubeconfig_path, "auth", "can-i", "--list")
			output, err := kubectlCmd.Output()
			if err != nil {
				log.Fatalln("error occured while fetching user access:", err)
			}
			fmt.Println("here are your current access authorities. please contact cluster owner if any is missing.")
			fmt.Println(string(output))
		},
	}

	// --kubeconfig flag
	initCmd.PersistentFlags().String("kubeconfig", util.Get_kubeconfig_path()+"config", "path to assign to the created kubeconfig file")

	// --userid flag ==> we use StringVarP to also have a shortened flag
	initCmd.PersistentFlags().StringVarP(&signum, "userid", "u", "", "ericsson signum")

	// --password flag ==> we use StringVarP to also have a shortened flag
	initCmd.PersistentFlags().StringVarP(&pass, "password", "p", "", "user password")

	// --static flag
	initCmd.PersistentFlags().Bool("static", false, "create static kubeconfig file")

	return initCmd
}

// This function returns the api server endpoint based on a given cluster name
func getAPIserverEndpoint(clusterName string) string {

	params := url.Values{
		"w":       {"ae"},
		"a":       {"f"},
		"cluster": {clusterName},
	}

	// TODO: add some timeout logic
	resp, err := http.PostForm("https://ews.rnd.gic.ericsson.se/a/", params)
	if err != nil {
		log.Fatalln("error in requesting API Server endpoint!")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	// check if results is correct
	if len(body) == 0 {
		log.Fatalln("could not retrieve the API server endpoint for the given cluster name!")
	}
	apiServerEndpoint := string(body)

	return apiServerEndpoint
}

// Creates a config file for kubectl using the given credentials and api server endpoint
// and saves it to the given path
func createKubectlConfig(apiServerEndpoint string, clusterName string, signum string, kubeconfig_path string) {

	data := struct {
		Signum            string
		APIserverEndpoint string
		ClusterName       string
	}{
		Signum:            signum,
		APIserverEndpoint: apiServerEndpoint,
		ClusterName:       clusterName,
	}

	var buf bytes.Buffer
	var err = kubectlConfigTemplate.Execute(&buf, &data)
	if err != nil {
		return
	}

	// Save the config file to YAML
	err = ioutil.WriteFile(kubeconfig_path, buf.Bytes(), 0600)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("user kubeconfig file has been saved in:", kubeconfig_path)
	fmt.Println("the kubeconfig file is only your identity for authentication, it does not mean you have cluster access.")
}

// Creates a static kubeconfig file given the necessary arguments and saves it
func createStaticConfig(apiServerEndpoint string,
	clusterName string,
	signum string,
	kubeconfig_path string,
	user_cert, user_key string) {

	user_cert = b64.StdEncoding.EncodeToString([]byte(user_cert))
	user_key = b64.StdEncoding.EncodeToString([]byte(user_key))

	data := struct {
		Signum            string
		APIserverEndpoint string
		ClusterName       string
		ClientCert        string
		ClientKey         string
	}{
		Signum:            signum,
		APIserverEndpoint: apiServerEndpoint,
		ClusterName:       clusterName,
		ClientCert:        user_cert,
		ClientKey:         user_key,
	}

	var buf bytes.Buffer
	var err = staticConfigTemplate.Execute(&buf, &data)
	if err != nil {
		return
	}

	// Save the config file to YAML
	err = ioutil.WriteFile(kubeconfig_path, buf.Bytes(), 0600)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("user static kubeconfig file has been saved in:", kubeconfig_path)
	fmt.Println("the kubeconfig file is only your identity for authentication, it does not mean you have cluster access.")
}
