/* Project Encore: BFG - Localized Private Game Restoration Server
 * Copyright (C) 2026 Paficent <paficent@tutamail.com> & Contributors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package game

// TODO: copied from python double check
var obstaclePosX = []int{14, 5, 9, 22, 7, 6, 16, 5, 7, 11, 13, 10, 20, 15, 12, 16, 22, 11, 26, 24, 18, 23, 3, 16, 2, 8, 14, 28, 6, 16, 13, 2, 22, 19, 11, 13, 28, 29, 22, 19, 6, 14, 26, 11, 12, 17, 21, 10, 7, 17, 1, 15, 9, 19, 25, 14, 21, 19, 10, 9}
var obstaclePosY = []int{12, 26, 21, 32, 17, 14, 27, 23, 27, 9, 21, 28, 35, 17, 31, 22, 26, 17, 31, 33, 29, 28, 20, 34, 23, 32, 33, 22, 31, 10, 23, 18, 23, 12, 13, 29, 20, 24, 9, 32, 21, 10, 22, 20, 34, 13, 25, 24, 9, 36, 20, 36, 9, 37, 21, 26, 29, 22, 14, 34}
var obstacleOffsets = []int{0, 5, 3, 4, 1, 2, 4, 3, 0, 5, 0, 2, 1, 3, 0, 1, 3, 4, 5, 0, 1, 2, 3, 0, 1, 3, 4, 0, 1, 2, 3, 5, 4, 0, 1, 1, 3, 1, 0, 1, 4, 0, 0, 2, 0, 3, 0, 0, 3, 4, 0, 3, 0, 0, 1, 3, 0, 0, 3, 1}
var obstacleBasePerIsland = []int{106, 112, 118, 125, 131, -1, 194, 208, -1, -1, -1, -1, 361}

func obstacleBase(islandID int64) int {
	idx := int(islandID) - 1
	if idx < 0 || idx >= len(obstacleBasePerIsland) {
		return -1
	}
	return obstacleBasePerIsland[idx]
}

func (p *Player) buildIsland(userIslandID, islandID, castle int64) *Island {
	now := nowMS()
	isl := &Island{UserIslandID: userIslandID, IslandID: islandID, BBBID: p.BBBID}
	add := func(structureID int64, x, y int) {
		isl.Structures = append(isl.Structures, &Structure{
			UserStructureID:   p.NextStructureID(),
			UserIslandID:      userIslandID,
			StructureID:       structureID,
			X:                 x,
			Y:                 y,
			Scale:             1,
			DateCreated:       now,
			LastCollection:    now,
			Muted:             0,
			IsComplete:        1,
			IsObstacleComplete:        1,
			IsUpgrading:       0,
			BuildingCompleted: now,
		})
	}
	add(castle, 29, 9)
	if islandID != 6 { // don't add breeding structure and nursery to gold island
		add(1, 35, 17)
		add(2, 21, 3)
	}
	if base := obstacleBase(islandID); base > 0 {
		for i := range obstaclePosX {
			add(int64(base+obstacleOffsets[i]), obstaclePosX[i], obstaclePosY[i])
		}
	}
	return isl
}

func (p *Player) NewIsland(islandID, castle int64) *Island {
	p.islandSeq++
	return p.buildIsland(p.islandSeq, islandID, castle)
}
