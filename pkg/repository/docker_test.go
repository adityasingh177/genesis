/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package repository

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	entityMock "github.com/whiteblock/genesis/mocks/pkg/entity"

	"github.com/docker/docker/api/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDockerRepository_GetNetworkByName_Success(t *testing.T) {
	results := []types.NetworkResource{
		types.NetworkResource{Name: "test1", ID: "id1"},
		types.NetworkResource{Name: "test2", ID: "id2"},
	}
	cli := new(entityMock.Client)
	cli.On("NetworkList", mock.Anything, mock.Anything).Return(results, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 2)
			assert.Nil(t, args.Get(0))
		}).Times(len(results) + 1)
	ds := NewDockerRepository(logrus.New())

	for _, result := range results {
		net, err := ds.GetNetworkByName(nil, cli, result.Name)
		assert.NoError(t, err)
		assert.Equal(t, result, net)
	}

	_, err := ds.GetNetworkByName(nil, cli, "DNE")
	assert.Error(t, err)

	cli.AssertExpectations(t)
}

func TestDockerRepository_GetNetworkByName_Failure(t *testing.T) {
	cli := new(entityMock.Client)
	cli.On("NetworkList", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("eerrr")).Once()
	ds := NewDockerRepository(logrus.New())
	_, err := ds.GetNetworkByName(nil, cli, "foo")
	assert.Error(t, err)

	cli.AssertExpectations(t)
}

func TestDockerRepository_HostHasImage_Success(t *testing.T) {
	testImageList := []types.ImageSummary{
		types.ImageSummary{RepoDigests: []string{"test0"}, RepoTags: []string{"test2"}},
		types.ImageSummary{RepoDigests: []string{"test3"}, RepoTags: []string{"test4"}},
		types.ImageSummary{RepoDigests: []string{"test5"}, RepoTags: []string{"test6"}},
		types.ImageSummary{RepoDigests: []string{"test7"}, RepoTags: []string{"test8"}},
	}

	existingImageTags := []string{"test2", "test6"}
	existingImageDigests := []string{"test0", "test3"}
	noneExistingImageTags := []string{"A", "B"}
	noneExistingImageDigests := []string{"C", "D"}

	cli := new(entityMock.Client)
	cli.On("ImageList", mock.Anything, mock.Anything).Return(testImageList, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 2)
			assert.Nil(t, args.Get(0))
		})

	ds := NewDockerRepository(logrus.New())

	for _, term := range append(existingImageTags, existingImageDigests...) {
		exists, err := ds.HostHasImage(nil, cli, term)
		assert.NoError(t, err)
		assert.True(t, exists)
	}

	for _, term := range append(noneExistingImageTags, noneExistingImageDigests...) {
		exists, err := ds.HostHasImage(nil, cli, term)
		assert.NoError(t, err)
		assert.False(t, exists)
	}
}

func TestDockerRepository_HostHasImage_Failure(t *testing.T) {

	cli := new(entityMock.Client)
	cli.On("ImageList", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("err"))

	ds := NewDockerRepository(logrus.New())
	exists, err := ds.HostHasImage(nil, cli, "foo")
	assert.Error(t, err)
	assert.False(t, exists)
}

func TestDockerRepository_EnsureImagePulled(t *testing.T) {
	testImageList := []types.ImageSummary{
		types.ImageSummary{RepoDigests: []string{"test0"}, RepoTags: []string{"test2"}},
		types.ImageSummary{RepoDigests: []string{"test3"}, RepoTags: []string{"test4"}},
		types.ImageSummary{RepoDigests: []string{"test5"}, RepoTags: []string{"test6"}},
		types.ImageSummary{RepoDigests: []string{"test7"}, RepoTags: []string{"test8"}},
	}

	existingImages := []string{"test7", "test6"}
	nonExistingImages := []string{"A", "B"}
	testReader := strings.NewReader("TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT")

	cli := new(entityMock.Client)
	cli.On("ImageList", mock.Anything, mock.Anything).Return(testImageList, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 2)
			assert.Nil(t, args.Get(0))
		}).Times(len(nonExistingImages) + len(existingImages))

	cli.On("ImagePull", mock.Anything, mock.Anything, mock.Anything).Return(
		ioutil.NopCloser(testReader), nil).Run(func(args mock.Arguments) {
		testReader.Reset("TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT")
		require.Len(t, args, 3)
		assert.Nil(t, args.Get(0))
		ipo, ok := args.Get(2).(types.ImagePullOptions)
		require.True(t, ok)
		assert.Equal(t, "Linux", ipo.Platform)
	}).Times(len(nonExistingImages))

	ds := NewDockerRepository(logrus.New())

	for _, img := range existingImages {
		err := ds.EnsureImagePulled(nil, cli, img, "")
		assert.NoError(t, err)
	}

	for _, img := range nonExistingImages {
		err := ds.EnsureImagePulled(nil, cli, img, "")
		assert.NoError(t, err)
	}
	cli.AssertExpectations(t)
}

func TestDockerRepository_EnsureImagePulled_ImagePull_Failure(t *testing.T) {
	testImageList := []types.ImageSummary{
		types.ImageSummary{RepoDigests: []string{"test0"}, RepoTags: []string{"test2"}},
		types.ImageSummary{RepoDigests: []string{"test3"}, RepoTags: []string{"test4"}},
	}

	cli := new(entityMock.Client)
	cli.On("ImageList", mock.Anything, mock.Anything, mock.Anything).Return(testImageList, nil).Once()

	cli.On("ImagePull", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		nil, fmt.Errorf("err")).Once()

	ds := NewDockerRepository(logrus.New())

	err := ds.EnsureImagePulled(nil, cli, "Foobar", "")
	assert.Error(t, err)
	cli.AssertExpectations(t)
}

func TestDockerRepository_GetContainerByName_Success(t *testing.T) {
	results := []types.Container{
		types.Container{Names: []string{"test1", "test3"}, ID: "id1"},
		types.Container{Names: []string{"test2", "test4"}, ID: "id2"},
	}
	cli := new(entityMock.Client)
	cli.On("ContainerList", mock.Anything, mock.Anything).Return(results, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 2)
			assert.Nil(t, args.Get(0))
		}).Times((2 * len(results)) + 1)
	ds := NewDockerRepository(logrus.New())

	for _, result := range results {
		for _, name := range result.Names {
			cntr, err := ds.GetContainerByName(nil, cli, name)
			assert.NoError(t, err)
			assert.Equal(t, result, cntr)
		}

	}

	_, err := ds.GetContainerByName(nil, cli, "DNE")
	assert.Error(t, err)

	cli.AssertExpectations(t)
}

func TestDockerRepository_GetContainerByName_Failure(t *testing.T) {
	cli := new(entityMock.Client)
	cli.On("ContainerList", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("err")).Once()
	ds := NewDockerRepository(logrus.New())
	_, err := ds.GetContainerByName(nil, cli, "DNE")
	assert.Error(t, err)

	cli.AssertExpectations(t)
}
