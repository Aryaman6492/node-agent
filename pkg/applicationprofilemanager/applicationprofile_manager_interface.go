package applicationprofilemanager

import (
	containercollection "github.com/inspektor-gadget/inspektor-gadget/pkg/container-collection"
	"github.com/Aryaman6492/node-agent/pkg/ebpf/events"
	tracerhardlinktype "github.com/Aryaman6492/node-agent/pkg/ebpf/gadgets/hardlink/types"
	tracerhttptype "github.com/Aryaman6492/node-agent/pkg/ebpf/gadgets/http/types"
	tracersymlinktype "github.com/Aryaman6492/node-agent/pkg/ebpf/gadgets/symlink/types"
	"github.com/Aryaman6492/node-agent/pkg/utils"
	"github.com/Aryaman6492/storage/pkg/apis/softwarecomposition/v1beta1"
)

type ApplicationProfileManagerClient interface {
	ContainerCallback(notif containercollection.PubSubEvent)
	RegisterPeekFunc(peek func(mntns uint64) ([]string, error))
	ReportCapability(k8sContainerID, capability string)
	ReportFileExec(k8sContainerID string, event events.ExecEvent)
	ReportFileOpen(k8sContainerID string, event events.OpenEvent)
	ReportHTTPEvent(k8sContainerID string, event *tracerhttptype.Event)
	ReportRulePolicy(k8sContainerID, ruleId, allowedProcess string, allowedContainer bool)
	ReportIdentifiedCallStack(k8sContainerID string, callStack *v1beta1.IdentifiedCallStack)
	ReportSymlinkEvent(k8sContainerID string, event *tracersymlinktype.Event)
	ReportHardlinkEvent(k8sContainerID string, event *tracerhardlinktype.Event)
	ReportDroppedEvent(k8sContainerID string)
	ContainerReachedMaxTime(containerID string)
}

type Enricher interface {
	EnrichEvent(k8sContainerID string, event utils.EnrichEvent, callID string)
}
