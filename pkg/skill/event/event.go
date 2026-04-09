/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package event

import (
	aiTypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	"github.com/pole-io/pole-server/pkg/common/eventhub"
)

// PublishSkillCreatedEvent publishes a skill created event
func PublishSkillCreatedEvent(skill *aiTypes.Skill) {
	_ = eventhub.Publish(eventhub.SkillEventTopic, &eventhub.SkillEvent{
		Skill:     skill,
		EventType: eventhub.EventCreated,
	})
}

// PublishSkillUpdatedEvent publishes a skill updated event
func PublishSkillUpdatedEvent(skill *aiTypes.Skill) {
	_ = eventhub.Publish(eventhub.SkillEventTopic, &eventhub.SkillEvent{
		Skill:     skill,
		EventType: eventhub.EventUpdated,
	})
}

// PublishSkillDeletedEvent publishes a skill deleted event
func PublishSkillDeletedEvent(skill *aiTypes.Skill) {
	_ = eventhub.Publish(eventhub.SkillEventTopic, &eventhub.SkillEvent{
		Skill:     skill,
		EventType: eventhub.EventDeleted,
	})
}

// PublishSkillVersionCreatedEvent publishes a skill version created event
func PublishSkillVersionCreatedEvent(version *aiTypes.SkillVersion) {
	_ = eventhub.Publish(eventhub.SkillVersionEventTopic, &eventhub.SkillVersionEvent{
		Version:   version,
		EventType: eventhub.EventCreated,
	})
}

// PublishSkillVersionActivatedEvent publishes a skill version activated event
func PublishSkillVersionActivatedEvent(version *aiTypes.SkillVersion) {
	_ = eventhub.Publish(eventhub.SkillVersionEventTopic, &eventhub.SkillVersionEvent{
		Version:   version,
		EventType: eventhub.EventUpdated,
	})
}

// PublishSkillSubscriptionCreatedEvent publishes a skill subscription created event
func PublishSkillSubscriptionCreatedEvent(sub *aiTypes.SkillSubscription) {
	_ = eventhub.Publish(eventhub.SkillSubscriptionEventTopic, &eventhub.SkillSubscriptionEvent{
		Subscription: sub,
		EventType:    eventhub.EventCreated,
	})
}

// PublishSkillSubscriptionDeletedEvent publishes a skill subscription deleted event
func PublishSkillSubscriptionDeletedEvent(sub *aiTypes.SkillSubscription) {
	_ = eventhub.Publish(eventhub.SkillSubscriptionEventTopic, &eventhub.SkillSubscriptionEvent{
		Subscription: sub,
		EventType:    eventhub.EventDeleted,
	})
}
