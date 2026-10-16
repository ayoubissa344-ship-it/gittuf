// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package gittuf

import (
	"testing"

	"github.com/gittuf/gittuf/pkg/gitinterface"
	"github.com/gittuf/gittuf/pkg/rsl"
	"github.com/stretchr/testify/require"
)

func TestSyncPushesReferenceUpdatedByPropagationEntry(t *testing.T) {
	const refName = "refs/heads/main"
	const upstreamRepository = "https://example.com/controller.git"

	remoteDir := t.TempDir()
	remoteR := gitinterface.CreateTestGitRepository(t, remoteDir, true)
	remoteRepo := &Repository{r: newStorer(remoteR)}

	baseBlobID, err := remoteR.WriteBlob([]byte("base"))
	require.NoError(t, err)
	baseTreeBuilder := gitinterface.NewTreeBuilder(remoteR)
	baseTreeID, err := baseTreeBuilder.WriteTreeFromEntries([]gitinterface.TreeEntry{
		gitinterface.NewEntryBlob("file.txt", baseBlobID),
	})
	require.NoError(t, err)

	baseCommitID, err := remoteRepo.r.Commit(baseTreeID, refName, "Base commit\n", false)
	require.NoError(t, err)
	require.NoError(t, rsl.NewReferenceEntry(refName, baseCommitID).Commit(remoteRepo.r, false))

	localDir := t.TempDir()
	localR, err := gitinterface.CloneAndFetchRepository(remoteDir, localDir, refName, []string{rsl.Ref}, true)
	require.NoError(t, err)
	require.NoError(t, localR.SetGitConfig("user.name", "Test Contributor"))
	require.NoError(t, localR.SetGitConfig("user.email", "test@example.com"))
	localRepo := &Repository{r: newStorer(localR)}

	newBlobID, err := localR.WriteBlob([]byte("propagated"))
	require.NoError(t, err)
	newTreeBuilder := gitinterface.NewTreeBuilder(localR)
	newTreeID, err := newTreeBuilder.WriteTreeFromEntries([]gitinterface.TreeEntry{
		gitinterface.NewEntryBlob("file.txt", newBlobID),
	})
	require.NoError(t, err)
	propagatedCommitID, err := localRepo.r.Commit(newTreeID, refName, "Propagated commit\n", false)
	require.NoError(t, err)

	baseEntry, err := rsl.GetLatestEntry(localRepo.r)
	require.NoError(t, err)
	propagationEntry := rsl.NewPropagationEntry(refName, propagatedCommitID, upstreamRepository, baseEntry.GetID())
	require.NoError(t, propagationEntry.Commit(localRepo.r, false))

	// Sync must push the target ref as well as the RSL entry that records it.
	divergedRefs, err := localRepo.sync("origin", false)
	require.NoError(t, err)
	require.Empty(t, divergedRefs)

	remoteTip, err := remoteR.GetReference(refName)
	require.NoError(t, err)
	require.Equal(t, propagatedCommitID, remoteTip,
		"the remote RSL must not advance without the reference target recorded by its PropagationEntry")

	remoteLatestEntry, err := rsl.GetLatestEntry(remoteR)
	require.NoError(t, err)
	remotePropagationEntry, ok := remoteLatestEntry.(*rsl.PropagationEntry)
	require.True(t, ok, "the propagated update must be present in the remote RSL")
	require.Equal(t, propagatedCommitID, remotePropagationEntry.GetTargetID())
}

func TestSyncPullsReferenceUpdatedByPropagationEntry(t *testing.T) {
	const refName = "refs/heads/main"
	const upstreamRepository = "https://example.com/controller.git"

	remoteDir := t.TempDir()
	remoteR := gitinterface.CreateTestGitRepository(t, remoteDir, true)
	remoteRepo := &Repository{r: newStorer(remoteR)}

	baseBlobID, err := remoteR.WriteBlob([]byte("base"))
	require.NoError(t, err)
	baseTreeBuilder := gitinterface.NewTreeBuilder(remoteR)
	baseTreeID, err := baseTreeBuilder.WriteTreeFromEntries([]gitinterface.TreeEntry{
		gitinterface.NewEntryBlob("file.txt", baseBlobID),
	})
	require.NoError(t, err)

	baseCommitID, err := remoteRepo.r.Commit(baseTreeID, refName, "Base commit\n", false)
	require.NoError(t, err)
	require.NoError(t, rsl.NewReferenceEntry(refName, baseCommitID).Commit(remoteRepo.r, false))
	baseEntry, err := rsl.GetLatestEntry(remoteRepo.r)
	require.NoError(t, err)

	localDir := t.TempDir()
	localR, err := gitinterface.CloneAndFetchRepository(remoteDir, localDir, refName, []string{rsl.Ref}, true)
	require.NoError(t, err)
	require.NoError(t, localR.SetGitConfig("user.name", "Test Contributor"))
	require.NoError(t, localR.SetGitConfig("user.email", "test@example.com"))
	localRepo := &Repository{r: newStorer(localR)}

	// Update the remote ref and record only a PropagationEntry for that update.
	newBlobID, err := remoteR.WriteBlob([]byte("propagated"))
	require.NoError(t, err)
	newTreeBuilder := gitinterface.NewTreeBuilder(remoteR)
	newTreeID, err := newTreeBuilder.WriteTreeFromEntries([]gitinterface.TreeEntry{
		gitinterface.NewEntryBlob("file.txt", newBlobID),
	})
	require.NoError(t, err)
	propagatedCommitID, err := remoteRepo.r.Commit(newTreeID, refName, "Propagated commit\n", false)
	require.NoError(t, err)
	require.NoError(t, rsl.NewPropagationEntry(refName, propagatedCommitID, upstreamRepository, baseEntry.GetID()).Commit(remoteRepo.r, false))

	// Sync must use the propagated target to update the local ref as well as RSL.
	divergedRefs, err := localRepo.sync("origin", false)
	require.NoError(t, err)
	require.Empty(t, divergedRefs)

	localTip, err := localR.GetReference(refName)
	require.NoError(t, err)
	require.Equal(t, propagatedCommitID, localTip,
		"the local ref must follow the target recorded by the remote PropagationEntry")
	assertLocalAndRemoteRefsMatch(t, localR, remoteR, rsl.Ref)
}
