/*
 * [y] hybris Platform
 *
 * Copyright (c) 2000-2018 hybris AG
 * All rights reserved.
 *
 * This software is the confidential and proprietary information of hybris
 * ("Confidential Information"). You shall not disclose such Confidential
 * Information and shall use it only in accordance with the terms of the
 * license agreement you entered into with hybris.
 */
package remoteenv

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServiceDefinitionService_NameNormalization(t *testing.T) {

	t.Run("should replace upper case with lower case", func(t *testing.T) {
		// when
		normalised := createServiceName("DisplayName", "id")

		// then
		assert.Equal(t, "displayname-87ea5", normalised)
	})

	t.Run("should replace non alpha numeric characters with --", func(t *testing.T) {
		// when
		normalised := createServiceName("display_!@#$%^&*()name", "id")

		// then
		assert.Equal(t, "display-name-87ea5", normalised)
	})

	t.Run("should remove leading dashes", func(t *testing.T) {
		// when
		normalised := createServiceName("-----displayname", "id")

		// then
		assert.Equal(t, "displayname-87ea5", normalised)
	})

	t.Run("should trim if name too long", func(t *testing.T) {
		// when
		normalised := createServiceName("VeryVeryVeryVeryVeryVeryVEryVeryVeryVeryVeryVeryVeryVeryLongDescription", "id")

		// then
		assert.Equal(t, "veryveryveryveryveryveryveryveryveryveryveryveryveryveryl-87ea5", normalised)
	})
}
